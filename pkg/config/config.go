package config

type Config struct {
	Port          int    `env:"PORT" envDefault:"8080"`
	LogLevel      string `env:"LOG_LEVEL" envDefault:"info"`
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"info"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:"info"`
}
