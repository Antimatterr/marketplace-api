package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Antimatterr/marketplace-api/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	args := os.Args
	fmt.Println(args)
	if len(args) < 2 {
		log.Fatal("usage: migrate <up | down>")
	}

	cfg := config.MustLoad()

	m, err := migrate.New(
		"file://migrations",
		cfg.DatabaseURL)

	if err != nil {
		log.Fatalf("migration.new: %v", err)
	}

	switch args[1] {
	case "up":
		log.Printf("migration up called")
		if err := m.Up(); err != nil {
			log.Fatalf("migration.up: %v", err)
		}
	case "down":
		log.Printf("migration down called")
		if err := m.Steps(-1); err != nil {
			log.Fatalf("migration.down: %v", err)
		}
	case "version":
		version, _, err := m.Version()
		if err != nil {
			log.Printf("error: %v", err)
		}
		log.Printf("Current migration: %d", version)

	default:
		log.Fatalf("unknowsn command for migration: %s", args[1])
	}

}
