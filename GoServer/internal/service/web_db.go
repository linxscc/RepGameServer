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

// GetDBConfig 根据环境返回数据库配置
func GetDBConfig() DBConfig {
	if os.Getenv("DOCKER_BUILD") == "1" {
		return DBConfig{
			Host:     "repgame-database-0.cx2omeoogidr.ap-southeast-2.rds.amazonaws.com",
			Port:     "3306",
			User:     "repgameadmin",
			Password: "repgameadmin",
			DBName:   "InsByMyself",
		}
	}
	return DBConfig{
		Host:     "127.0.0.1",
		Port:     "13306",
		User:     "repgameadmin",
		Password: "repgameadmin",
		DBName:   "InsByMyself",
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
