CREATE DATABASE IF NOT EXISTS RepGame CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
GRANT ALL PRIVILEGES ON RepGame.* TO 'repgameadmin'@'%';

USE RepGame;

CREATE TABLE IF NOT EXISTS UserAccount (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ResponseInfo (
    id INT PRIMARY KEY,
    code VARCHAR(32) NOT NULL,
    response_key VARCHAR(100) NOT NULL,
    message VARCHAR(255) DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO ResponseInfo (id, code, response_key, message) VALUES
    (1001, '1001', 'Connected', 'Connected'),
    (2001, '2001', 'LoginSuccess', 'Login successful'),
    (2002, '2002', 'InvalidRequest', 'Invalid request'),
    (2003, '2003', 'UserNotFound', 'User not found'),
    (2004, '2004', 'LoginFailed', 'Login failed'),
    (2005, '2005', 'AlreadyLoggedIn', 'User already logged in'),
    (3001, '3001', 'RegisterSuccess', 'Registration successful'),
    (3002, '3002', 'InvalidRegisterRequest', 'Invalid registration request'),
    (3003, '3003', 'InvalidCredentials', 'Invalid username or password'),
    (3004, '3004', 'UserExists', 'User already exists'),
    (3005, '3005', 'RegisterFailed', 'Registration failed'),
    (4002, '4002', 'InvalidGameState', 'Invalid game state'),
    (4003, '4003', 'PlayerNotFound', 'Player not found'),
    (5001, '5001', 'GameStarted', 'Game started'),
    (5002, '5002', 'BondList', 'Bond list'),
    (5009, '5009', 'MissingCard', 'Card information is required'),
    (6001, '6001', 'Reconnected', 'Reconnected'),
    (6002, '6002', 'ReconnectFailed', 'Reconnect failed'),
    (7001, '7001', 'PlayerDisconnected', 'Player disconnected'),
    (7002, '7002', 'PlayerReconnected', 'Player reconnected'),
    (9999, '9999', 'UnknownError', 'Unknown request');

CREATE TABLE IF NOT EXISTS CardDeck (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    cards_num INT NOT NULL DEFAULT 1,
    damage DECIMAL(10,2) NOT NULL DEFAULT 0,
    targetname VARCHAR(100) DEFAULT NULL,
    level INT NOT NULL DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO CardDeck (id, name, cards_num, damage, targetname, level) VALUES
    (1, 'Fire', 10, 10, NULL, 1),
    (2, 'Water', 10, 10, NULL, 1),
    (3, 'Wind', 10, 10, NULL, 1);

CREATE TABLE IF NOT EXISTS Bonds (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    level INT NOT NULL,
    damage DECIMAL(10,2) NOT NULL DEFAULT 0,
    skill VARCHAR(100) DEFAULT '',
    description VARCHAR(255) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS BondCards (
    id INT AUTO_INCREMENT PRIMARY KEY,
    bond_id INT NOT NULL UNIQUE,
    card_name1 VARCHAR(100),
    card_name2 VARCHAR(100),
    card_name3 VARCHAR(100),
    card_name4 VARCHAR(100),
    card_name5 VARCHAR(100),
    card_name6 VARCHAR(100),
    card_name7 VARCHAR(100),
    CONSTRAINT fk_bond_cards_bond FOREIGN KEY (bond_id) REFERENCES Bonds(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

FLUSH PRIVILEGES;
