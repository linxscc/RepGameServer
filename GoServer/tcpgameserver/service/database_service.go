package service

import (
	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/models"
	"database/sql"
	"fmt"
	"log"

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
		log.Printf("Failed to connect to database: %v", err)
		return nil, err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Printf("Failed to ping database: %v", err)
		db.Close()
		return nil, err
	}

	return db, nil
}

// GetAllResponseInfo 获取所有响应信息
func GetAllResponseInfo() ([]models.ResponseInfo, error) {
	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := "SELECT id, code, response_key, COALESCE(message, '') as message FROM ResponseInfo ORDER BY id"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query ResponseInfo: %v", err)
	}
	defer rows.Close()

	var responseInfos []models.ResponseInfo
	for rows.Next() {
		var info models.ResponseInfo
		err := rows.Scan(&info.ID, &info.Code, &info.ResponseKey, &info.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ResponseInfo: %v", err)
		}
		responseInfos = append(responseInfos, info)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during ResponseInfo iteration: %v", err)
	}

	return responseInfos, nil
}

// GetAllCardDeck 获取所有卡牌信息
func GetAllCardDeck() ([]models.CardDeck, error) {
	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := "SELECT id, name, cards_num, damage, targetname, level FROM CardDeck ORDER BY level, id"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query CardDeck: %v", err)
	}
	defer rows.Close()

	var cardDecks []models.CardDeck
	for rows.Next() {
		var card models.CardDeck
		var targetname sql.NullString

		err := rows.Scan(&card.ID, &card.Name, &card.CardsNum, &card.Damage, &targetname, &card.Level)
		if err != nil {
			return nil, fmt.Errorf("failed to scan CardDeck: %v", err)
		}

		// 处理可能为NULL的targetname字段
		if targetname.Valid {
			card.TargetName = &targetname.String
		} else {
			card.TargetName = nil
		}

		cardDecks = append(cardDecks, card)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during CardDeck iteration: %v", err)
	}

	return cardDecks, nil
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

	log.Printf("Retrieved user account: %s", username)
	return &user, nil
}

// ValidateUserLogin 验证用户登录
func ValidateUserLogin(username, password string) (bool, error) {
	user, err := GetUserAccount(username)
	if err != nil {
		return false, err
	}

	// 这里应该使用加密密码比较，暂时使用明文比较
	// 在生产环境中应该使用 bcrypt 等加密方式
	if user.Password == password {
		log.Printf("User login successful: %s", username)
		return true, nil
	}

	return false, nil
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

	// 插入新用户
	query := "INSERT INTO UserAccount (username, password) VALUES (?, ?)"
	_, err = db.Exec(query, username, password)
	if err != nil {
		return fmt.Errorf("failed to create user account: %v", err)
	}
	log.Printf("User account created successfully: %s", username)
	return nil
}

// GetAllBonds 获取所有羁绊数据（包含关联的卡牌）
func GetAllBonds() ([]models.BondModel, error) {
	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// 查询羁绊基本信息
	bondsQuery := `
		SELECT id, name, level, damage, 
		       COALESCE(skill, '') as skill, 
		       description 
		FROM Bonds ORDER BY id`

	bondRows, err := db.Query(bondsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query bonds: %v", err)
	}
	defer bondRows.Close()
	var bonds []models.BondModel
	bondIndexMap := make(map[int]int)
	// 读取羁绊基本信息
	for bondRows.Next() {
		var bond models.BondModel
		err := bondRows.Scan(
			&bond.ID,
			&bond.Name,
			&bond.Level,
			&bond.Damage,
			&bond.Skill,
			&bond.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bond: %v", err)
		}

		// 初始化卡牌切片
		bond.CardNames = make([]string, 0)
		bonds = append(bonds, bond)
		bondIndexMap[bond.ID] = len(bonds) - 1
	}
	// 查询羁绊关联的卡牌
	cardQuery := `
		SELECT br.bond_id, br.card_name1, br.card_name2, br.card_name3, 
		       br.card_name4, br.card_name5, br.card_name6, br.card_name7
		FROM BondCards br
		ORDER BY br.bond_id`
	cardRows, err := db.Query(cardQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query BondCards: %v", err)
	}
	defer cardRows.Close() // 读取羁绊关联的卡牌信息
	for cardRows.Next() {
		var bondID int
		var cardName1, cardName2, cardName3, cardName4, cardName5, cardName6, cardName7 sql.NullString
		err := cardRows.Scan(&bondID, &cardName1, &cardName2, &cardName3, &cardName4, &cardName5, &cardName6, &cardName7)
		if err != nil {
			return nil, fmt.Errorf("failed to scan BondCards: %v", err)
		}

		// 将非空的卡牌名称添加到对应的羁绊中
		if index, exists := bondIndexMap[bondID]; exists {
			cardNames := []sql.NullString{cardName1, cardName2, cardName3, cardName4, cardName5, cardName6, cardName7}
			for _, cardName := range cardNames {
				if cardName.Valid && cardName.String != "" {
					bonds[index].CardNames = append(bonds[index].CardNames, cardName.String)
				}
			}
		} else {
			log.Printf("Warning: Bond ID %d not found in bonds", bondID)
		}
	}
	return bonds, nil
}
