package service

import (
	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/models"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // MySQL驱动
)

// GetDBConnection 获取数据库连接
func GetDBConnection() (*sql.DB, error) {
	dbConfig := config.GetDBConfig()

	// 构建MySQL连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// CheckUserAccountExists 检查用户账户是否存在
func CheckUserAccountExists(username string) (bool, error) {
	db, err := GetDBConnection()
	if err != nil {
		return false, err
	}
	defer db.Close()

	var count int
	query := "SELECT COUNT(*) FROM UserAccount WHERE username = ?"
	err = db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %v", err)
	}

	return count > 0, nil
}

// GetUserAccount 获取用户账户信息（用于验证登录）
func GetUserAccount(username string) (*models.UserAccount, error) {
	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var user models.UserAccount
	query := "SELECT username, password FROM UserAccount WHERE username = ?"
	err = db.QueryRow(query, username).Scan(&user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, fmt.Errorf("failed to get user account: %v", err)
	}

	return &user, nil
}

// CreateUserAccount 创建新用户账户
func CreateUserAccount(username, password string) error {
	db, err := GetDBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	// 检查用户是否已存在
	exists, err := CheckUserAccountExists(username)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user already exists: %s", username)
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	// 插入新用户（只保存 bcrypt 哈希）
	query := "INSERT INTO UserAccount (username, password) VALUES (?, ?)"
	_, err = db.Exec(query, username, hash)
	if err != nil {
		return fmt.Errorf("failed to create user account: %v", err)
	}
	return nil
}
