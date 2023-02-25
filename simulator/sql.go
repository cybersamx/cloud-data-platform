package main

import (
	"database/sql"
	"embed"
	"log"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	dsn                 = "host=localhost port=5433 user=pguser password=password dbname=db_test sslmode=disable"
	codeUniqueViolation = "23505"
	dirMigrations       = "migrations"
)

//go:embed migrations/*.sql
var initSQLDir embed.FS

func connectDB() (*sql.DB, error) {
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

func initDB() error {
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

		log.Println(string(buf))
	}

	return nil
}
