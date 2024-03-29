package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bensolo-io/aws-redis-simple/pkg/config"
	"github.com/redis/go-redis/v9"

	"github.com/caarlos0/env/v7"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	cfg    config.Config
	rdb    redis.UniversalClient
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	// red    = color.FgRed.Render
	// green  = color.FgGreen.Render
	// yellow = color.FgYellow.Render
)

func init() {
	gin.SetMode(gin.ReleaseMode)
	cfg = config.Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(-1)
	}
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	color.NoColor = cfg.LogNoColor
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: cfg.LogNoColor})
	redisAddr := fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort)
	redisUOpts := &redis.UniversalOptions{
		Addrs:    []string{redisAddr},
		DB:       cfg.RedisDbIndex,
		Password: cfg.RedisPassword,
		Username: "",
		// empty tls.Config is required to enable TLS - uses OS cert chain for verification
		TLSConfig: &tls.Config{
			InsecureSkipVerify: cfg.RedisInsecureSkipVerify,
		},
	}

	log.Warn().Msgf("connecting with address %s, using password %t", redisAddr, cfg.RedisPassword != "")

	// see https://github.com/golang/go/issues/51991
	// TLDR: mac os forces SCT validation, amazon wont provide SCT in certs
	if strings.EqualFold(runtime.GOOS, "darwin") {
		log.Warn().Msgf("%s for more info see https://github.com/golang/go/issues/51991", yellow("detected darwin runtime - disabling tls verification;"))
		redisUOpts.TLSConfig.InsecureSkipVerify = true
	}

	log.Info().Msgf("%+v", redisUOpts)

	rdb = redis.NewUniversalClient(redisUOpts)
	log.Info().Msgf("redis client created for %s", redisAddr)

	// initial redis check for pod readiness
	// func() {
	// 	tReady := time.NewTicker(time.Second * 5)
	// 	for {
	// 		if err := checkRedis(context.Background(), "-readiness"); err != nil {
	// 			log.Error().Msgf(red("failed initial redis check, trying again in 5 seconds: %s"), err)
	// 		} else {
	// 			return
	// 		}
	// 		<-tReady.C
	// 	}
	// }()

	// log.Info().Msg(green("initial redis check OK"))

	ginEngine := gin.New()
	ginEngine.Use(gin.Recovery())
	// if we make it this far the initial check passed, so just return 200
	ginEngine.GET("/readiness", func(ctx *gin.Context) { ctx.String(200, "ok") })
	ginEngine.GET("/liveness", func(ctx *gin.Context) {
		if err := checkRedis(ctx, "-liveness"); err != nil {
			log.Error().Msgf(red("failed redis check: %s"), err.Error())
			// ctx.String(500, err.Error())
			// return
		}
		log.Debug().Msg(green("/liveness OK"))
		ctx.String(200, "ok")
	})

	startGinServerWithGracefulShutdown(ginEngine, fmt.Sprintf(":%d", cfg.Port))
}

func checkRedis(ctx context.Context, keySuffix string) error {
	now := time.Now().String()
	key := fmt.Sprintf("%s%s", cfg.RedisTestKeyName, keySuffix)

	err := rdb.Set(ctx, key, now, 0).Err()
	if err != nil {
		return fmt.Errorf("could not set key %s: %s", key, err)
	}

	redisNow, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("could not get key %s: %s", key, err)
	}

	if !strings.EqualFold(now, redisNow) {
		return fmt.Errorf("redis operation succeeded for key %s but values do not match: %s != %s", key, now, redisNow)
	}

	return nil
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
