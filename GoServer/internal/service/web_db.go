package service

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// DBConfig 数据库配置（与 tcpgameserver 完全解耦）
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// GetDBConfig 从环境变量读取数据库配置
func GetDBConfig() DBConfig {
	return DBConfig{
		Host:     envOrDefault("DB_HOST", "127.0.0.1"),
		Port:     envOrDefault("DB_PORT", "13306"),
		User:     envOrDefault("DB_USER", "repgameadmin"),
		Password: envOrDefault("DB_PASSWORD", "repgameadmin"),
		DBName:   envOrDefault("WEB_DB_NAME", "InsByMyself"),
	}
}

// GetDB 获取数据库连接
func GetDB() (*sql.DB, error) {
	cfg := GetDBConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("[web_db] Failed to open database: %v", err)
		return nil, err
	}
	if err := db.Ping(); err != nil {
		log.Printf("[web_db] Failed to ping database: %v", err)
		db.Close()
		return nil, err
	}
	return db, nil
}
