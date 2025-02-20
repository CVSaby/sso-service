package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env      string        `yaml:"env" env-default:"local"`
	TokenTTL time.Duration `yaml:"token_ttl" env-default:"1h"`
	GRPC     GRPCConfig    `yaml:"grpc" env-required:"true"`
	JWT      JWTConfig     `yaml:"jwt" env-required:"true"`
	DBConfig DBConfig      `yaml:"db" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-default:"4000"`
	Timeout time.Duration `yaml:"timeout"`
}

type JWTConfig struct {
	Secret              string        `yaml:"secret_string" env-required:"true"`
	AccessTokenLifeTime time.Duration `yaml:"access_token_life_time" env-default:"1h"`
}

type DBConfig struct {
	Host   string `yaml:"host" env-required:"true"`
	Port   int    `yaml:"port" env-default:"5432"`
	DBName string `yaml:"db_name" env-required:"true"`
	DBUser string `yaml:"db_user" env-required:"true"`
	DBPass string `yaml:"db_password" env-required:"true"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
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

// fetchConfigPath fetches config path from command line flag or env variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
