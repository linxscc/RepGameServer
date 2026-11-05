CREATE DATABASE IF NOT EXISTS RepGame CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
GRANT ALL PRIVILEGES ON RepGame.* TO 'repgameadmin'@'%';

USE RepGame;

CREATE TABLE IF NOT EXISTS UserAccount (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 卡牌、羁绊、消息文本等游戏数据不再存放在数据库中，
-- 统一由服务器配置表管理：GoServer/tcpgameserver/config/tables/*.yaml
-- 数据库只保存玩家账号。

FLUSH PRIVILEGES;
