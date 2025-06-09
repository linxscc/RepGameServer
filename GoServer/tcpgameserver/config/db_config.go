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

// GetDBConfig 根据环境返回数据库配置
func GetDBConfig() DBConfig {

	// 检查是否在Docker环境中运行
	if os.Getenv("DOCKER_BUILD") == "1" {
		// Docker环境下使用RDS连接配置
		return DBConfig{
			Host:     "repgame-database-0.cx2omeoogidr.ap-southeast-2.rds.amazonaws.com",
			Port:     "3306",
			User:     "repgameadmin",
			Password: "repgameadmin",
			DBName:   "RepGame",
		}
	}

	// 本地开发环境配置
	return DBConfig{
		Host:     "127.0.0.1",
		Port:     "13306",
		User:     "repgameadmin",
		Password: "repgameadmin",
		DBName:   "RepGame",
	}
}
