package main

import (
	"errors"
	"fmt"
	migratorconfig "github.com/CVSaby/sso-service/cmd/migrator/config"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

//postgres://username:password@localhost:5432/dbname?sslmode=disable

func main() {
	log.Println("Starting migration...")
	cfg, migrationsPath, migrationsTable := migratorconfig.MustLoad()

	migAddress := fmt.Sprintf(
		"postgres://%s:%s@%s:%v/%s?sslmode=disable&x-migrations-table=%s",
		cfg.DBConfig.DBUser,
		cfg.DBConfig.DBPass,
		cfg.DBConfig.Host,
		cfg.DBConfig.Port,
		cfg.DBConfig.DBName,
		migrationsTable,
	)

	m, err := migrate.New("file://"+migrationsPath, migAddress)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No migrations to apply")
			return
		}

		panic(err)
	}

	fmt.Println("Applied migrations")
}
