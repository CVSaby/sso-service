package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env string `yaml:"env" env-default:"prod" env:"env" envDefault:"prod"`

	ServiceName    string `yaml:"service_name" env:"SERVICE_NAME,required"`
	ServiceVersion string `yaml:"service_version" env:"SERVICE_VERSION,required"`

	OTLPEndpoint string `yaml:"otlp_endpoint" env:"OTLP_ENDPOINT,required"`

	GRPC        GRPCConfig  `yaml:"grpc" env-required:"true"`
	KafkaConfig KafkaConfig `yaml:"kafka" env-required:"true"`

	JWT      JWTConfig `yaml:"jwt" env-required:"true"`
	DBConfig DBConfig  `yaml:"db" env-required:"true"`
}

type KafkaConfig struct {
	Servers  []string `yaml:"servers" env-required:"true" env:"KAFKA_SERVERS,required"`
	Topic    string   `yaml:"topic" env-required:"true" env:"KAFKA_TOPIC,required"`
	ClientID string   `yaml:"client_id" env-required:"true" env:"KAFKA_CLIENT_ID,required"`
}

type LoggerConfig struct {
	LoggerAddr string `yaml:"logger_addr" env-required:"true" env:"LOGGER_ADDR"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-default:"4000" env:"GRPC_PORT,required"`
	Timeout time.Duration `yaml:"timeout" env:"GRPC_TIMEOUT" envDefault:"5s"`
}

type JWTConfig struct {
	Secret              string        `yaml:"secret_string" env-required:"true" env:"JWT_SECRET,required"`
	AccessTokenLifeTime time.Duration `yaml:"access_token_life_time" env-default:"1h" env:"JWT_ACCESS_TOKEN_LIFE_TIME,required"`
}

type DBConfig struct {
	Host   string `yaml:"host" env-required:"true" env:"DB_HOST,required"`
	Port   int    `yaml:"port" env-default:"5432" env:"DB_PORT,required"`
	DBName string `yaml:"db_name" env-required:"true" env:"DB_NAME,required"`
	DBUser string `yaml:"db_user" env-required:"true" env:"DB_USER,required"`
	DBPass string `yaml:"db_password" env-required:"true" env:"DB_PASSWORD,required"`
}

func MustLoad() *Config {
	cfgType := os.Getenv("CONFIG_TYPE")
	if cfgType == "" {
		panic("config_type environment variable not set")
	}

	switch cfgType {
	case "file":
		return loadConfigFromFile()
	case "env":
		return loadConfigFromEnv()
	default:
		return loadConfigFromEnv()
	}
}

func loadConfigFromFile() *Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("Failed to read config: " + err.Error())
	}

	return &cfg
}

func loadConfigFromEnv() *Config {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		panic("Failed to parse config: " + err.Error())
	}

	cfg, err = env.ParseAs[Config]()
	if err != nil {
		panic("Failed to parse config: " + err.Error())
	}

	return &cfg
}
