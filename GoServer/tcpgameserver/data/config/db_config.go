package config

import (
	"os"
	"strconv"
	"time"
)

// DBConfig 数据库配置信息（全部来自环境变量，见 docker-compose.yml）
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string

	MaxOpenConns    int           // 连接池最大连接数（DB_MAX_OPEN_CONNS）
	MaxIdleConns    int           // 连接池最大空闲连接数（DB_MAX_IDLE_CONNS）
	ConnMaxLifetime time.Duration // 单个连接最长使用时间（DB_CONN_MAX_LIFETIME_MINUTES）
	QueryTimeout    time.Duration // 单次查询超时（DB_QUERY_TIMEOUT_SECONDS）
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v > 0 {
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
		DBName:   envOrDefault("DB_NAME", "RepGame"),

		MaxOpenConns:    envIntOrDefault("DB_MAX_OPEN_CONNS", 10),
		MaxIdleConns:    envIntOrDefault("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: time.Duration(envIntOrDefault("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute,
		QueryTimeout:    time.Duration(envIntOrDefault("DB_QUERY_TIMEOUT_SECONDS", 5)) * time.Second,
	}
}
