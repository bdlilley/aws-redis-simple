package config

type Config struct {
	Port                    int    `env:"PORT" envDefault:"8080"`
	LogLevel                string `env:"LOG_LEVEL" envDefault:"info"`
	RedisAddr               string `env:"REDIS_ADDR,required"`
	RedisDbIndex            int    `env:"REDIS_DB_INDEX" envDefault:"0"`
	RedisPassword           string `env:"REDIS_PASSWORD,required"`
	RedisTestKeyName        string `env:"REDIS_TEST_KEY_NAME" envDefault:"local"`
	RedisInsecureSkipVerify bool   `env:"REDIS_INSECURE_SKIP_VERIFY" envDefault:"false"`
}
