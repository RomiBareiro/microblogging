package config

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func ConnectDB() *sql.DB {
	dsn := "postgres://postgres:password@localhost:5432/blog?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("DB unreachable: %v", err)
	}

	return db
}
