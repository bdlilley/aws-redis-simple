package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bensolo-io/aws-redis-simple/pkg/config"
	"github.com/redis/go-redis/v9"

	"github.com/caarlos0/env/v7"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var cfg config.Config

func init() {
	gin.SetMode(gin.ReleaseMode)
	cfg = config.Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		// DB index 0 is required for elasticache
		DB:       0,
		Password: cfg.RedisPassword,
		TLSConfig: &tls.Config{
			// see https://github.com/golang/go/issues/51991
			// TLDR: mac os forcing SCT validation, amazon wont provide SCT in certs
			InsecureSkipVerify: true,
		},
	})
	log.Info().Msgf("connected to redis ")

	ginEngine := gin.New()
	ginEngine.Use(gin.Recovery())
	ginEngine.GET("/", func(ctx *gin.Context) {
		now := time.Now().String()

		err := rdb.Set(ctx, "aws_redis_sample_now", now, 0).Err()
		if err != nil {
			err = fmt.Errorf("could not set key: %s", err)
			log.Error().Err(err).Msg("")
			ctx.String(500, err.Error())
			return
		}

		redisNow, err := rdb.Get(ctx, "aws_redis_sample_now").Result()
		if err != nil {
			err = fmt.Errorf("could not get key: %s", err)
			log.Error().Err(err).Msg("")
			ctx.String(500, err.Error())
			return
		}

		if !strings.EqualFold(now, redisNow) {
			err = fmt.Errorf("redis operation succeeded but values do not match: %s != %s", now, redisNow)
			log.Error().Err(err).Msg("")
			ctx.String(500, err.Error())
			return
		}

		ctx.String(200, fmt.Sprintf("ok %s", redisNow))
	})

	startGinServerWithGracefulShutdown(ginEngine, fmt.Sprintf(":%d", cfg.Port))
}

func startGinServerWithGracefulShutdown(r *gin.Engine, listenerAddr string) {
	srv := &http.Server{
		Addr:    listenerAddr,
		Handler: r,
	}

	go func() {
		log.Info().Msgf("starting container server on %s", listenerAddr)
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg(err.Error())
		}
	}()
	quit := make(chan os.Signal, 10)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Warn().Msg("shutting down due to signal")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Msg("forced shutdown, timeout exceeded")
	}

	log.Warn().Msg("shutdown complete")
}
