package tcpserver

import (
	"errors"
	"log"
	"strings"

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// HandleLogin 用户名 + 密码登录。
func HandleLogin(ctx *Context, msg *pb.ClientMessage) {
	req := msg.GetLogin()
	if err := service.ValidateCredentialFormat(req.GetUsername(), req.GetPassword()); err != nil {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_CREDENTIALS, "invalid username or password")
		return
	}

	if err := service.AuthenticateUser(req.GetUsername(), req.GetPassword()); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_CREDENTIALS, "invalid username or password")
		} else {
			log.Printf("login %q: %v", req.GetUsername(), err)
			ctx.Fail(pb.ErrorCode_ERROR_CODE_INTERNAL, "login temporarily unavailable")
		}
		return
	}

	completeAuthentication(ctx, req.GetUsername())
}

// HandleResumeSession 断线重连时用令牌恢复身份。
func HandleResumeSession(ctx *Context, msg *pb.ClientMessage) {
	username, ok := service.GetSessionStore().Resolve(msg.GetResumeSession().GetSessionToken())
	if !ok {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_SESSION_EXPIRED, "session expired, please log in again")
		return
	}
	completeAuthentication(ctx, username)
}

// HandleRegister 注册新账号。
func HandleRegister(ctx *Context, msg *pb.ClientMessage) {
	req := msg.GetRegister()
	if err := service.ValidateCredentialFormat(req.GetUsername(), req.GetPassword()); err != nil {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_REQUEST, err.Error())
		return
	}

	if err := service.CreateUserAccount(req.GetUsername(), req.GetPassword()); err != nil {
		if strings.Contains(err.Error(), "user already exists") {
			ctx.Fail(pb.ErrorCode_ERROR_CODE_USER_EXISTS, "username already taken")
		} else {
			log.Printf("register %q: %v", req.GetUsername(), err)
			ctx.Fail(pb.ErrorCode_ERROR_CODE_INTERNAL, "registration failed")
		}
		return
	}

	ctx.Reply(&pb.ServerMessage{Body: &pb.ServerMessage_RegisterResult{
		RegisterResult: &pb.RegisterResult{Username: req.GetUsername()},
	}})
}

// completeAuthentication 身份已确认后的统一流程：处理重复登录/断线重连、
// 绑定连接、签发会话令牌并回复 AuthResult。
func completeAuthentication(ctx *Context, username string) {
	connManager := ctx.ConnManager
	resumingGame := false

	if existing, loggedIn := connManager.GetConnectionByUsername(username); loggedIn && existing.ClientID != ctx.ClientID {
		if existing.GetStatus() == types.StatusWaitingReconnect {
			resumingGame = true
		} else {
			// 同一账号在别处在线：踢掉旧连接（同步执行，保证随后绑定的是新连接）
			kickData := events.CreateUserConnectionEventData(
				events.EventClientKicked, existing.ClientID, username, existing.RemoteAddr)
			kickData.AddData("kick_reason", "duplicate_login")
			kickData.AddData("kicked_by", "system")
			kickData.AddData("new_client_id", ctx.ClientID)
			events.PublishSync(events.EventClientKicked, kickData)
		}
	}

	token, expiresAt, err := service.GetSessionStore().Issue(username)
	if err != nil {
		log.Printf("issue session for %q: %v", username, err)
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INTERNAL, "login temporarily unavailable")
		return
	}

	if !resumingGame {
		if err := connManager.BindUser(ctx.ClientID, username); err != nil {
			log.Printf("bind %q to %s: %v", username, ctx.ClientID, err)
			ctx.Fail(pb.ErrorCode_ERROR_CODE_INTERNAL, "login temporarily unavailable")
			return
		}
		connManager.SetPlayerStatus(ctx.ClientID, types.StatusLoggedIn)
	}

	ctx.Reply(&pb.ServerMessage{Body: &pb.ServerMessage_AuthResult{AuthResult: &pb.AuthResult{
		Username:           username,
		SessionToken:       token,
		SessionExpiresAtMs: expiresAt.UnixMilli(),
	}}})

	if resumingGame {
		// 由重连处理器绑定连接并推送当前对局状态
		reconnectData := events.CreateUserConnectionEventData(
			events.EventClientReconnect, ctx.ClientID, username, ctx.Conn.RemoteAddr().String())
		events.Publish(events.EventClientReconnect, reconnectData)
	}
}
