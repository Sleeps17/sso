package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

// TODO: подумать над миграциями: избавиться от колонки is_admin и попробовать перенести ее в отдельную таблицу

func main() {
	var storagePath, migrationsPath, migrationsTable string

	flag.StringVar(&storagePath, "storage-path", "", "path to storage")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")

	flag.Parse()

	if storagePath == "" {
		panic("storage path is required")
	}

	if migrationsPath == "" {
		panic("migrations path is required")
	}

	migrator, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("sqlite3://%s", storagePath),
	)
	if err != nil {
		panic(err)
	}

	if err := migrator.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("no migrations to apply")

			return
		}
		panic(err)
	}

	log.Println("migrations applied successfully")
}
