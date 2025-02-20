package migratorconfig

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	DBConfig DBConfig `yaml:"db" env-required:"true"`
}

type DBConfig struct {
	Host   string `yaml:"host" env-required:"true"`
	Port   int    `yaml:"port" env-default:"5432"`
	DBName string `yaml:"db_name" env-required:"true"`
	DBUser string `yaml:"db_user" env-required:"true"`
	DBPass string `yaml:"db_password" env-required:"true"`
}

func MustLoad() (config *Config, mPath string, mTable string) {
	cfgPath, migPath, migTable := fetchPaths()
	if cfgPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		panic("config file does not exist: " + cfgPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		panic("Failed to read config: " + err.Error())
	}

	return &cfg, migPath, migTable
}

// fetchPaths fetches config path, migrations path, migrations table from command line flag or env variable.
// Default value of migrations table is empty migrations.
func fetchPaths() (cfgPath string, migPath string, migTable string) {
	var cfg, migrationsPath, migrationsTable string

	flag.StringVar(&migrationsPath, "migrations-path", "", "Path to migrations folder")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.StringVar(&cfg, "config", "", "path to config file")
	flag.Parse()

	if migrationsPath == "" {
		migrationsPath = os.Getenv("MIGRATIONS_PATH")
	}

	if cfg == "" {
		cfg = os.Getenv("CONFIG_PATH")
	}

	return cfg, migrationsPath, migrationsTable
}
