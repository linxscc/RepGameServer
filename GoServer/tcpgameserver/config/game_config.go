package config

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"GoServer/tcpgameserver/rules"
)

// 配置表文件随代码一起编译进二进制（config/tables/*.yaml），
// 如设置了环境变量 GAME_CONFIG_DIR，则优先读取该目录下的同名文件，
// 方便在不重新编译的情况下调整数值。

//go:embed tables/*.yaml
var embeddedTables embed.FS

// GameRules 对局规则（game_rules.yaml）
type GameRules struct {
	Match struct {
		PlayersPerRoom int `yaml:"players_per_room"` // 每个房间的玩家数量（含机器人）
	} `yaml:"match"`
	Player struct {
		InitialHealth   float64 `yaml:"initial_health"`    // 初始血量（同时是血量上限）
		InitialHandSize int     `yaml:"initial_hand_size"` // 开局手牌数
		MaxHandCards    int     `yaml:"max_hand_cards"`    // 手牌上限
		DrawAfterPlay   int     `yaml:"draw_after_play"`   // 出牌后补牌数
	} `yaml:"player"`
	Turn struct {
		TimeoutSeconds int `yaml:"timeout_seconds"` // 人类玩家每回合思考时间
	} `yaml:"turn"`
	Compose struct {
		GroupSize int `yaml:"group_size"` // 合成所需同名卡牌数
	} `yaml:"compose"`
}

// TurnTimeout 人类玩家回合超时时间
func (r GameRules) TurnTimeout() time.Duration {
	return time.Duration(r.Turn.TimeoutSeconds) * time.Second
}

// BotConfig 机器人对局（bot.yaml）
type BotConfig struct {
	Enabled          bool   `yaml:"enabled"`            // 是否启用机器人对局
	MatchWaitSeconds int    `yaml:"match_wait_seconds"` // 等待真人对手的最长时间
	NamePrefix       string `yaml:"name_prefix"`        // 机器人名称前缀（禁止注册）
	ThinkDelayMs     int    `yaml:"think_delay_ms"`     // 机器人出牌前的思考时间
	MaxCardsPerPlay  int    `yaml:"max_cards_per_play"` // 机器人一次最多出牌数
}

// MatchWait 开始匹配后等待真人对手的时长
func (b BotConfig) MatchWait() time.Duration {
	return time.Duration(b.MatchWaitSeconds) * time.Second
}

// ThinkDelay 机器人出牌前的思考时长
func (b BotConfig) ThinkDelay() time.Duration {
	return time.Duration(b.ThinkDelayMs) * time.Millisecond
}

// NetworkConfig 网络与账号安全（network.yaml）
type NetworkConfig struct {
	Connection struct {
		ListenAddr              string `yaml:"listen_addr"`               // 监听地址
		MaxFrameBytes           int    `yaml:"max_frame_bytes"`           // 单帧最大字节数
		HandshakeTimeoutSeconds int    `yaml:"handshake_timeout_seconds"` // TLS 握手超时
		IdleTimeoutSeconds      int    `yaml:"idle_timeout_seconds"`      // 无消息断线时间
		StaleConnectionSeconds  int    `yaml:"stale_connection_seconds"`  // 兜底清理：视为失效的无活动时长
		CleanupIntervalMinutes  int    `yaml:"cleanup_interval_minutes"`  // 兜底清理的检查间隔
		WriteTimeoutSeconds     int    `yaml:"write_timeout_seconds"`     // 单条消息写超时
		SendQueueSize           int    `yaml:"send_queue_size"`           // 每连接待发送队列长度
		ReconnectGraceSeconds   int    `yaml:"reconnect_grace_seconds"`   // 对局中断线后等待重连的时长
	} `yaml:"connection"`
	Auth struct {
		BcryptCost        int `yaml:"bcrypt_cost"`         // bcrypt 强度
		SessionTTLHours   int `yaml:"session_ttl_hours"`   // 会话令牌有效期
		UsernameMinLength int `yaml:"username_min_length"` // 用户名最短字符数
		UsernameMaxLength int `yaml:"username_max_length"` // 用户名最长字符数
		PasswordMinLength int `yaml:"password_min_length"` // 密码最短字节数
		PasswordMaxLength int `yaml:"password_max_length"` // 密码最长字节数

		MaxConcurrentHashes           int `yaml:"max_concurrent_hashes"`              // 并发密码哈希上限，0=CPU 核数
		LoginAttemptsPerIPPerMinute   int `yaml:"login_attempts_per_ip_per_minute"`   // 每 IP 每分钟登录/注册次数
		LoginAttemptsPerUserPerMinute int `yaml:"login_attempts_per_user_per_minute"` // 每用户名每分钟登录次数
	} `yaml:"auth"`
}

// HandshakeTimeout TLS 握手超时
func (n NetworkConfig) HandshakeTimeout() time.Duration {
	return time.Duration(n.Connection.HandshakeTimeoutSeconds) * time.Second
}

// IdleTimeout 无消息断线时间
func (n NetworkConfig) IdleTimeout() time.Duration {
	return time.Duration(n.Connection.IdleTimeoutSeconds) * time.Second
}

// StaleConnection 连接记录视为失效的无活动时长
func (n NetworkConfig) StaleConnection() time.Duration {
	return time.Duration(n.Connection.StaleConnectionSeconds) * time.Second
}

// CleanupInterval 连接兜底清理的检查间隔
func (n NetworkConfig) CleanupInterval() time.Duration {
	return time.Duration(n.Connection.CleanupIntervalMinutes) * time.Minute
}

// WriteTimeout 单条消息写超时
func (n NetworkConfig) WriteTimeout() time.Duration {
	return time.Duration(n.Connection.WriteTimeoutSeconds) * time.Second
}

// ReconnectGrace 对局中断线后等待重连的时长
func (n NetworkConfig) ReconnectGrace() time.Duration {
	return time.Duration(n.Connection.ReconnectGraceSeconds) * time.Second
}

// SessionTTL 会话令牌有效期
func (n NetworkConfig) SessionTTL() time.Duration {
	return time.Duration(n.Auth.SessionTTLHours) * time.Hour
}

// Tables 全部配置表
type Tables struct {
	Rules    GameRules
	Bot      BotConfig
	Network  NetworkConfig
	Messages Messages
	Cards    CardTable // 卡牌数据（cards.yaml）
	Bonds    BondTable // 羁绊数据（bonds.yaml）

	// Chess 新版规则（chess_*.yaml），由 rules 包使用
	Chess *rules.Config
}

var (
	loadOnce sync.Once
	tables   *Tables
)

// Game 返回全局配置表；首次调用时加载，配置非法时直接终止进程（启动即失败，便于发现问题）。
func Game() *Tables {
	loadOnce.Do(func() {
		t, err := Load(os.Getenv("GAME_CONFIG_DIR"))
		if err != nil {
			log.Fatalf("load game config tables: %v", err)
		}
		tables = t
	})
	return tables
}

// Load 读取并校验配置表；overrideDir 为空时只使用内置配置。
func Load(overrideDir string) (*Tables, error) {
	t := &Tables{}
	var (
		chessRules    rules.RulesTable
		chessLines    rules.LinesFile
		chessBonds    rules.BondsFile
		chessRelics   rules.RelicsFile
		chessFamilies rules.FamiliesFile
	)
	files := []struct {
		name string
		dst  interface{}
	}{
		{"game_rules.yaml", &t.Rules},
		{"bot.yaml", &t.Bot},
		{"network.yaml", &t.Network},
		{"messages.yaml", &t.Messages},
		{"cards.yaml", &t.Cards},
		{"bonds.yaml", &t.Bonds},
		{"chess_rules.yaml", &chessRules},
		{"chess_lines.yaml", &chessLines},
		{"chess_bonds.yaml", &chessBonds},
		{"chess_relics.yaml", &chessRelics},
		{"chess_families.yaml", &chessFamilies},
	}
	for _, f := range files {
		data, err := readTable(overrideDir, f.name)
		if err != nil {
			return nil, err
		}
		dec := yaml.NewDecoder(bytesReader(data))
		dec.KnownFields(true) // 拼错的字段名直接报错，而不是被静默忽略
		if err := dec.Decode(f.dst); err != nil {
			return nil, fmt.Errorf("%s: %w", f.name, err)
		}
	}
	chess, err := rules.NewConfig(chessRules, chessLines, chessBonds, chessRelics, chessFamilies)
	if err != nil {
		return nil, err
	}
	t.Chess = chess
	if err := t.validate(); err != nil {
		return nil, err
	}
	return t, nil
}

func readTable(overrideDir, name string) ([]byte, error) {
	if overrideDir != "" {
		data, err := os.ReadFile(filepath.Join(overrideDir, name))
		if err == nil {
			return data, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
	}
	return embeddedTables.ReadFile("tables/" + name)
}

// validate 检查数值是否在合理范围内
func (t *Tables) validate() error {
	var errs []error
	check := func(ok bool, format string, args ...interface{}) {
		if !ok {
			errs = append(errs, fmt.Errorf(format, args...))
		}
	}

	r := t.Rules
	check(r.Match.PlayersPerRoom == 2, "game_rules.match.players_per_room must be 2 (turn logic is 1v1), got %d", r.Match.PlayersPerRoom)
	check(r.Player.InitialHealth > 0, "game_rules.player.initial_health must be > 0")
	check(r.Player.InitialHandSize > 0 && r.Player.InitialHandSize <= r.Player.MaxHandCards,
		"game_rules.player.initial_hand_size must be in 1..max_hand_cards")
	check(r.Player.DrawAfterPlay >= 0, "game_rules.player.draw_after_play must be >= 0")
	check(r.Turn.TimeoutSeconds >= 5, "game_rules.turn.timeout_seconds must be >= 5")
	check(r.Compose.GroupSize >= 2, "game_rules.compose.group_size must be >= 2")

	b := t.Bot
	check(b.MatchWaitSeconds >= 0, "bot.match_wait_seconds must be >= 0")
	check(b.NamePrefix != "", "bot.name_prefix must not be empty")
	check(b.ThinkDelayMs >= 0, "bot.think_delay_ms must be >= 0")
	check(b.MaxCardsPerPlay >= 1, "bot.max_cards_per_play must be >= 1")

	n := t.Network
	check(n.Connection.ListenAddr != "", "network.connection.listen_addr must not be empty")
	check(n.Connection.MaxFrameBytes >= 1024, "network.connection.max_frame_bytes must be >= 1024")
	check(n.Connection.HandshakeTimeoutSeconds > 0, "network.connection.handshake_timeout_seconds must be > 0")
	check(n.Connection.IdleTimeoutSeconds >= 20, "network.connection.idle_timeout_seconds must be >= 20 (client heartbeat is 15s)")
	check(n.Connection.StaleConnectionSeconds > 0, "network.connection.stale_connection_seconds must be > 0")
	check(n.Connection.CleanupIntervalMinutes > 0, "network.connection.cleanup_interval_minutes must be > 0")
	check(n.Connection.WriteTimeoutSeconds > 0, "network.connection.write_timeout_seconds must be > 0")
	check(n.Connection.SendQueueSize >= 8, "network.connection.send_queue_size must be >= 8")
	check(n.Connection.ReconnectGraceSeconds > 0, "network.connection.reconnect_grace_seconds must be > 0")
	check(n.Auth.MaxConcurrentHashes >= 0, "network.auth.max_concurrent_hashes must be >= 0")
	check(n.Auth.LoginAttemptsPerIPPerMinute > 0, "network.auth.login_attempts_per_ip_per_minute must be > 0")
	check(n.Auth.LoginAttemptsPerUserPerMinute > 0, "network.auth.login_attempts_per_user_per_minute must be > 0")
	check(n.Auth.BcryptCost >= 10 && n.Auth.BcryptCost <= 14, "network.auth.bcrypt_cost must be 10..14")
	check(n.Auth.SessionTTLHours > 0, "network.auth.session_ttl_hours must be > 0")
	check(n.Auth.UsernameMinLength > 0 && n.Auth.UsernameMinLength <= n.Auth.UsernameMaxLength,
		"network.auth username length range is invalid")
	check(n.Auth.PasswordMinLength > 0 && n.Auth.PasswordMinLength <= n.Auth.PasswordMaxLength && n.Auth.PasswordMaxLength <= 72,
		"network.auth password length range is invalid (max 72)")

	if err := t.Messages.validate(); err != nil {
		errs = append(errs, err)
	}
	if err := validateCardData(t.Cards, t.Bonds); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func bytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }
