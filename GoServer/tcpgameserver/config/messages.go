package config

import (
	"fmt"
	"sort"
)

// Messages 服务器消息文本表（messages.yaml），键 -> 文本。
type Messages map[string]string

// 消息键。代码只通过这些常量引用文本，启动时会校验它们都在表中。
const (
	MsgMalformedMessage   = "malformed_message"
	MsgEmptyMessageBody   = "empty_message_body"
	MsgUnsupportedRequest = "unsupported_request"

	MsgLoginRequired      = "login_required"
	MsgInvalidCredentials = "invalid_credentials"
	MsgLoginUnavailable   = "login_unavailable"
	MsgSessionExpired     = "session_expired"
	MsgLoggedInElsewhere  = "logged_in_elsewhere"
	MsgUsernameTaken      = "username_taken"
	MsgRegisterFailed     = "register_failed"
	MsgUsernameLength     = "username_length"
	MsgUsernameSpaces     = "username_spaces"
	MsgUsernameReserved   = "username_reserved"
	MsgPasswordLength     = "password_length"

	MsgNoCardsToPlay       = "no_cards_to_play"
	MsgComposeCountInvalid = "compose_count_invalid"
	MsgCardsInvalid        = "cards_invalid"
	MsgNotYourTurn         = "not_your_turn"

	MsgReconnectPlayerNotFound = "reconnect_player_not_found"
	MsgReconnectNotWaiting     = "reconnect_not_waiting"
	MsgReconnectNoGame         = "reconnect_no_game"
	MsgReconnectConnectionLost = "reconnect_connection_lost"
	MsgReconnectBindFailed     = "reconnect_bind_failed"
)

var requiredMessageKeys = []string{
	MsgMalformedMessage, MsgEmptyMessageBody, MsgUnsupportedRequest,
	MsgLoginRequired, MsgInvalidCredentials, MsgLoginUnavailable, MsgSessionExpired,
	MsgLoggedInElsewhere, MsgUsernameTaken, MsgRegisterFailed, MsgUsernameLength,
	MsgUsernameSpaces, MsgUsernameReserved, MsgPasswordLength,
	MsgNoCardsToPlay, MsgComposeCountInvalid, MsgCardsInvalid, MsgNotYourTurn,
	MsgReconnectPlayerNotFound, MsgReconnectNotWaiting, MsgReconnectNoGame,
	MsgReconnectConnectionLost, MsgReconnectBindFailed,
}

// Text 返回消息文本；args 非空时按占位符格式化。
func (m Messages) Text(key string, args ...interface{}) string {
	text, ok := m[key]
	if !ok {
		return key // 理论上不会发生（启动时已校验），兜底返回键名
	}
	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}
	return text
}

func (m Messages) validate() error {
	var missing []string
	for _, k := range requiredMessageKeys {
		if m[k] == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("messages.yaml is missing keys: %v", missing)
	}
	return nil
}

// Msg 是 config.Game().Messages.Text 的简写。
func Msg(key string, args ...interface{}) string {
	return Game().Messages.Text(key, args...)
}
