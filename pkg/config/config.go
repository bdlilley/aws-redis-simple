package config

type Config struct {
	Port                    int    `env:"PORT" envDefault:"8080"`
	LogLevel                string `env:"LOG_LEVEL" envDefault:"info"`
	LogNoColor              bool   `env:"LOG_NO_COLOR" envDefault:"false"`
	RedisHost               string `env:"REDIS_HOST,required"`
	RedisPort               int    `env:"REDIS_PORT" envDefault:"6397"`
	RedisDbIndex            int    `env:"REDIS_DB_INDEX" envDefault:"0"`
	RedisPassword           string `env:"REDIS_PASSWORD,required"`
	RedisTestKeyName        string `env:"REDIS_TEST_KEY_NAME" envDefault:"local"`
	RedisInsecureSkipVerify bool   `env:"REDIS_INSECURE_SKIP_VERIFY" envDefault:"false"`
}
