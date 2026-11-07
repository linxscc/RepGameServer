// Package store 账号数据的 MySQL 存取（数据层，不含业务规则）。
package store

import (
	"GoServer/tcpgameserver/data/config"
	"GoServer/tcpgameserver/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/go-sql-driver/mysql" // MySQL驱动
)

// ErrUserNotFound 账号不存在
var ErrUserNotFound = errors.New("user not found")

// ErrUserExists 注册时用户名已被占用
var ErrUserExists = errors.New("user already exists")

// mysqlDuplicateEntry MySQL 错误码 ER_DUP_ENTRY：违反唯一索引（用户名重复）
const mysqlDuplicateEntry = 1062

// 进程内共享一个连接池（*sql.DB 并发安全，内部复用连接）。
// sql.Open 不会立即建立连接，数据库暂时不可用时查询返回错误，恢复后自动重连。
var (
	dbOnce sync.Once
	dbPool *sql.DB
	dbErr  error
	dbCfg  config.DBConfig
)

func database() (*sql.DB, error) {
	dbOnce.Do(func() {
		dbCfg = config.GetDBConfig()
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.DBName)
		dbPool, dbErr = sql.Open("mysql", dsn)
		if dbErr != nil {
			return
		}
		dbPool.SetMaxOpenConns(dbCfg.MaxOpenConns)
		dbPool.SetMaxIdleConns(dbCfg.MaxIdleConns)
		dbPool.SetConnMaxLifetime(dbCfg.ConnMaxLifetime)
	})
	return dbPool, dbErr
}

// queryContext 每次查询都带超时，数据库卡住时不会无限占用连接协程
func queryContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), dbCfg.QueryTimeout)
}

// GetUserAccount 获取用户账户信息（用于验证登录）；账号不存在时返回 ErrUserNotFound
func GetUserAccount(username string) (*model.UserAccount, error) {
	db, err := database()
	if err != nil {
		return nil, err
	}
	ctx, cancel := queryContext()
	defer cancel()

	var user model.UserAccount
	err = db.QueryRowContext(ctx, "SELECT username, password FROM UserAccount WHERE username = ?", username).
		Scan(&user.Username, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user account: %w", err)
	}
	return &user, nil
}

// InsertUser 新建账号，passwordHash 必须已是 bcrypt 哈希；用户名已存在时返回 ErrUserExists
func InsertUser(username, passwordHash string) error {
	db, err := database()
	if err != nil {
		return err
	}
	ctx, cancel := queryContext()
	defer cancel()

	// 依赖 username 唯一索引判断重名，没有"先查再插"的竞态
	_, err = db.ExecContext(ctx, "INSERT INTO UserAccount (username, password) VALUES (?, ?)", username, passwordHash)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntry {
		return ErrUserExists
	}
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// UpdatePasswordHash 把旧的明文密码升级为 bcrypt 哈希
func UpdatePasswordHash(username, hash string) error {
	db, err := database()
	if err != nil {
		return err
	}
	ctx, cancel := queryContext()
	defer cancel()
	_, err = db.ExecContext(ctx, "UPDATE UserAccount SET password = ? WHERE username = ?", hash, username)
	return err
}
