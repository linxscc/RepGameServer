package service

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var voyaraDB *sql.DB

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func InitDB() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		envOrDefault("DB_USER", "repgameadmin"),
		envOrDefault("DB_PASSWORD", "repgameadmin"),
		envOrDefault("DB_HOST", "127.0.0.1"),
		envOrDefault("DB_PORT", "13306"),
		envOrDefault("VOYARA_DB_NAME", "Voyara"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("[voyara] Failed to open database: %v", err)
		return err
	}

	if err := db.Ping(); err != nil {
		log.Printf("[voyara] Failed to ping database: %v", err)
		db.Close()
		return err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	voyaraDB = db
	log.Printf("[voyara] Database connection pool initialized (max_open=25, max_idle=10)")
	return nil
}

func GetDB() (*sql.DB, error) {
	if voyaraDB == nil {
		return nil, fmt.Errorf("database not initialized, call InitDB() first")
	}
	return voyaraDB, nil
}

func CloseDB() error {
	if voyaraDB == nil {
		return nil
	}
	return voyaraDB.Close()
}
