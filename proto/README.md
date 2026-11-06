# RepGame 通信协议

`repgame/v1/game.proto` 是客户端与服务器协议的**唯一事实来源**，Go 和 C# 代码都由它生成。

## 分层

| 层 | 服务器 (Go) | 客户端 (Unity C#) |
|---|---|---|
| 传输 | `tls.Listen`，`server.go` | `SecureConnection`（TLS + 证书指纹固定） |
| 帧 | `protocol/frame.go` | `Net/FrameCodec.cs` |
| 编解码 | `protocol/codec.go` | `Google.Protobuf` |
| 分发 | `messageHandle/handler.go` 路由表 | `Net/ServerMessageRouter.cs` |
| 模型转换 | `protocol/mapper.go` | `Net/ProtocolMapper.cs` |
| 业务 | `logic/`、`service/` | UI 只订阅 `NetEvents` + 领域模型 |

帧格式：`[4 字节大端序长度][protobuf 消息]`，单帧上限 1MB。

## 安全约定

- **传输加密**：默认 TLS 1.2+。服务器首次启动自动生成自签名证书（Docker 卷 `repgame_game_tls_certs`），
  日志打印 SHA-256 指纹；客户端在 `.env` 的 `SERVER_CERT_SHA256` 固定该指纹，指纹不符直接拒绝连接。
- **密码**：只在 TLS 内传输；服务器用 bcrypt (cost 12) 存储。旧的明文密码在用户下次登录成功时自动升级为哈希。
  用户不存在与密码错误返回同一错误码，并做等时比较，避免枚举账号。
- **会话**：登录成功返回 `session_token`（32 字节随机数，24 小时有效，重新登录会作废旧令牌），
  断线重连用 `ResumeSessionRequest` 恢复身份，不再重发密码。
- **服务器权威**：客户端只发送出牌/合成的卡牌 UID，牌面数值一律取服务器手牌数据。
- **鉴权**：路由表中标记 `requireAuth` 的请求在未登录时统一返回 `ERROR_CODE_NOT_AUTHENTICATED`。
- **心跳**：客户端每 15 秒发送 `Ping`，服务器 45 秒收不到任何消息即断开。

## 修改协议

1. 只追加字段/分支，不修改或复用已有字段编号；删除字段用 `reserved`。
2. 不兼容的变更需递增 `protocol.ProtocolVersion`（通过 `Welcome` 下发）。
3. 重新生成代码：

   ```bash
   PROTOC=/path/to/protoc ./scripts/gen-proto.sh ../RepGame
   ```

4. 新增请求：在 `ClientMessage.body` 加分支 → 在 `messageHandle/handler.go` 的 `routes` 注册处理器。
   新增推送：在 `ServerMessage.body` 加分支 → 在客户端 `ServerMessageRouter` 构造函数注册。
