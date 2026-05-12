package config

import "os"

// DBConfig 数据库配置信息
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
		DBName:   envOrDefault("DB_NAME", "RepGame"),
	}
}
