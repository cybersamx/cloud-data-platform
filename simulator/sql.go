package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
)

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

func initDB(db *sql.DB, cfg config) error {
	for _, table := range cfg.Tables {
		var createStmt string

		switch table.Source.Type {
		case "csv":
			var colDefs []string
			for _, column := range table.Columns {
				colDefs = append(colDefs, fmt.Sprintf("%s %s NULL", column.Name, column.DataType))
			}

			createStmt = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ( %s )", table.Name, strings.Join(colDefs, ","))
		case "json":
			createStmt = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ( json TEXT )", table.Name)
		default:
			return errors.New("missing or invalid table source type")
		}

		_, err := db.Exec(createStmt)
		if err != nil {
			return err
		}
	}

	return nil
}
