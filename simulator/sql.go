package main

import (
	"database/sql"
	"embed"
	"log"
	"path/filepath"

	"github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	dirMigrations = "migrations"
)

//go:embed migrations/*.sql
var initSQLDir embed.FS

func stmtBuilder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

func connectDB(dsn string) (*sql.DB, error) {
	log.Println("Connecting to database.")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func initDB(db *sql.DB) error {
	log.Println("Running migrations.")

	files, err := initSQLDir.ReadDir(dirMigrations)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := filepath.Join(dirMigrations, file.Name())
		log.Printf("Migration: Running %s.", filename)

		buf, err := initSQLDir.ReadFile(filename)
		if err != nil {
			return err
		}

		_, err = db.Exec(string(buf))
		if err != nil {
			return err
		}
	}

	return nil
}
