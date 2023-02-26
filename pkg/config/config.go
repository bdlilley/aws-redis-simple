package config

type Config struct {
	Port             int    `env:"PORT" envDefault:"8080"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"info"`
	RedisAddr        string `env:"REDIS_ADDR,required"`
	RedisPassword    string `env:"REDIS_PASSWORD,required"`
	RedisTestKeyName string `env:"REDIS_TEST_KEY_NAME,required"`
}
