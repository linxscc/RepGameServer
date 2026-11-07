# RepGame 通信协议

`repgame/v1/*.proto` 是客户端与服务器协议的**唯一事实来源**，Go 和 C# 代码都由它生成。
按模块拆分为多个文件，同属 `package repgame.v1`（线上格式与拆分前完全一致）：

| 文件 | 内容 |
|---|---|
| `envelope.proto` | `ClientMessage` / `ServerMessage` 信封（oneof 分支决定消息类型） |
| `common.proto` | 错误码、错误、心跳 |
| `auth.proto` | 欢迎、登录、注册、恢复会话 |
| `lobby.proto` | 匹配与在线状态 |
| `match_action.proto` | 对局操作意图（`GameAction` 及各操作） |
| `match_config.proto` | 开局下发一次的静态数据（卡牌、羁绊、法宝、家世、天象） |
| `match_state.proto` | 对局动态状态（阶段、棋子、计分、回合结算、结束） |

生成物：Go 在 `GoServer/tcpgameserver/pb/<模块>.pb.go`，C# 在客户端
`Assets/Scripts/Protocol/Generated/<模块>.cs`。生成文件里的 `descriptorData` 是 protoc 编译出的
schema，protobuf 运行时靠它解析消息，不可手改、也不能挪到配置里。

## 分层

| 层 | 服务器 (Go) | 客户端 (Unity C#) |
|---|---|---|
| 传输 | `tls.Listen`，`server.go` | `SecureConnection`（TLS + 证书指纹固定） |
| 帧 | `network/transport/frame.go` | `Network/Transport/FrameCodec.cs` |
| 编解码 / 发送队列 | `network/transport/codec.go`、`conn_writer.go` | `Google.Protobuf`、`SecureConnection` |
| 分发 | `network/handler/handler.go` 路由表 | `Network/ServerMessageRouter.cs` |
| 模型转换 | `logic/match/chess_views.go` | `Network/ProtocolMapper.cs` |
| 业务 | `logic/`、`service/` | `Logic/`、`Service/`；UI 只订阅 `NetEvents` |

服务器目录分层（`GoServer/tcpgameserver/`）：

| 目录 | 职责 |
|---|---|
| `network/`（`transport`、`handler`、`tlsutil`） | 网络通信：帧、发送队列、TLS、请求路由与处理 |
| `model/` | 连接、房间、账号等数据结构 |
| `data/`（`config`、`store`） | 配置表加载（`config/tables/*.yaml`）、MySQL 存取 |
| `service/` | 账号认证、会话、连接管理、房间管理、限流 |
| `logic/`（`match`、`rules`、`chessbot`） | 对局流程、规则引擎（`rules/keys.go` 为配置 key 常量）、机器人 |

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

4. 新增请求：在 `envelope.proto` 的 `ClientMessage.body` 加分支 → 在 `network/handler/handler.go` 的 `routes` 注册处理器。
   新模块的消息放进对应模块的 `.proto`；新增 `.proto` 文件时同时加进 `scripts/gen-proto.sh` 的 `PROTOS` 列表。
   新增推送：在 `ServerMessage.body` 加分支 → 在客户端 `ServerMessageRouter` 构造函数注册。
