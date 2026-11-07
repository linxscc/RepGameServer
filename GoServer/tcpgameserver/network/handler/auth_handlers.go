package handler

import (
	"errors"
	"log"

	"GoServer/tcpgameserver/data/config"
	"GoServer/tcpgameserver/logic/match"
	"GoServer/tcpgameserver/model"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/service"
)

// HandleLogin 用户名 + 密码登录。
func HandleLogin(ctx *Context, msg *pb.ClientMessage) {
	req := msg.GetLogin()
	if !service.AllowAuthAttempt(ctx.RemoteIP(), req.GetUsername()) {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INTERNAL, config.Msg(config.MsgTooManyAttempts))
		return
	}
	if err := service.ValidateCredentialFormat(req.GetUsername(), req.GetPassword()); err != nil {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_CREDENTIALS, config.Msg(config.MsgInvalidCredentials))
		return
	}

	if err := service.AuthenticateUser(req.GetUsername(), req.GetPassword()); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_CREDENTIALS, config.Msg(config.MsgInvalidCredentials))
		} else {
			log.Printf("login %q: %v", req.GetUsername(), err)
			ctx.Fail(pb.ErrorCode_ERROR_CODE_INTERNAL, config.Msg(config.MsgLoginUnavailable))
		}
		return
	}

	completeAuthentication(ctx, req.GetUsername())
}

// HandleResumeSession 断线重连时用令牌恢复身份。
func HandleResumeSession(ctx *Context, msg *pb.ClientMessage) {
	username, ok := service.GetSessionStore().Resolve(msg.GetResumeSession().GetSessionToken())
	if !ok {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_SESSION_EXPIRED, config.Msg(config.MsgSessionExpired))
		return
	}
	completeAuthentication(ctx, username)
}

// HandleRegister 注册新账号。
func HandleRegister(ctx *Context, msg *pb.ClientMessage) {
	req := msg.GetRegister()
	if !service.AllowAuthAttempt(ctx.RemoteIP(), "") {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INTERNAL, config.Msg(config.MsgTooManyAttempts))
		return
	}
	if err := service.ValidateCredentialFormat(req.GetUsername(), req.GetPassword()); err != nil {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_REQUEST, err.Error())
		return
	}
	if service.IsReservedUsername(req.GetUsername()) {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_REQUEST, config.Msg(config.MsgUsernameReserved, config.Game().Bot.NamePrefix))
		return
	}

	if err := service.CreateUserAccount(req.GetUsername(), req.GetPassword()); err != nil {
		if errors.Is(err, service.ErrUserExists) {
			ctx.Fail(pb.ErrorCode_ERROR_CODE_USER_EXISTS, config.Msg(config.MsgUsernameTaken))
		} else {
			log.Printf("register %q: %v", req.GetUsername(), err)
			ctx.Fail(pb.ErrorCode_ERROR_CODE_INTERNAL, config.Msg(config.MsgRegisterFailed))
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

	if existing, loggedIn := connManager.GetConnectionByUsername(username); loggedIn && existing.ClientID != ctx.ClientID &&
		existing.GetStatus() != model.StatusWaitingReconnect {
		// 同一账号在别处在线：踢掉旧连接（对局中的旧连接会转为等待重连，由本连接恢复）
		match.KickDuplicateLogin(existing)
	}

	token, expiresAt, err := service.GetSessionStore().Issue(username)
	if err != nil {
		log.Printf("issue session for %q: %v", username, err)
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INTERNAL, config.Msg(config.MsgLoginUnavailable))
		return
	}

	resuming := match.CanResumeGame(username)
	if !resuming {
		if err := connManager.BindUser(ctx.ClientID, username); err != nil {
			log.Printf("bind %q to %s: %v", username, ctx.ClientID, err)
			ctx.Fail(pb.ErrorCode_ERROR_CODE_INTERNAL, config.Msg(config.MsgLoginUnavailable))
			return
		}
		connManager.SetPlayerStatus(ctx.ClientID, model.StatusLoggedIn)
	}

	ctx.Reply(&pb.ServerMessage{Body: &pb.ServerMessage_AuthResult{AuthResult: &pb.AuthResult{
		Username:           username,
		SessionToken:       token,
		SessionExpiresAtMs: expiresAt.UnixMilli(),
	}}})

	// 有对局可恢复：绑定到本连接并推送完整状态（在 AuthResult 之后到达客户端）
	if resuming && !match.ResumeGame(ctx.ClientID, username) {
		// 对局恰好在这期间结束：按普通登录处理
		if err := connManager.BindUser(ctx.ClientID, username); err == nil {
			connManager.SetPlayerStatus(ctx.ClientID, model.StatusLoggedIn)
		}
	}
}
