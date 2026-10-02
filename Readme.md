# RepGame TCP Server

`RepServer` 是专门维护游戏后端服务器的分支。它只包含：

- Go TCP 游戏服务器：`GoServer/tcpgameserver`
- 独立游戏服务器入口：`GoServer/main.go`
- MySQL 8.4 与 RepGame 初始化结构
- 可选的 Namecheap DDNS 更新服务
- 未来游戏后台管理界面的空目录：`Front/game-admin`

原 Web 网站、商城前后端、Nginx、HTTPS 和 Web API 均不包含在此分支。

## 启动

```bash
cp .env.example .env
```

至少修改 `.env` 中的：

- `DB_PASSWORD`
- `MYSQL_ROOT_PASSWORD`

启动游戏服务器和数据库：

```bash
docker compose up -d --build
docker compose ps
```

游戏客户端连接：

```text
本机：127.0.0.1:19060
局域网：服务器的局域网 IP:19060
外网：zsdimain.site:19060
```

路由器应转发 TCP `19060 → 服务器的局域网 IP:19060`。同一 Wi-Fi 内的客户端直接使用局域网 IP，不经过路由器端口转发。本分支不提供 Web 服务，因此端口 `80` 和 `443` 不再使用。

Compose 项目名固定为 `repgame-game-server`，使用独立的容器、镜像、网络和 `repgame_game_mysql_data` 数据卷。主机端口 `19060` 和 `23306` 也与网站版分开，因此两套 Docker 服务可以同时运行。

## 动态 DNS

如果 `.env` 已配置 `NAMECHEAP_DDNS_PASSWORD`：

```bash
docker compose --profile ddns up -d
docker compose --profile ddns logs -f ddns
```

## 数据库

MySQL 仅绑定本机 `127.0.0.1:23306`，游戏服务器通过 Docker 内部网络连接。数据保存在 `repgame_game_mysql_data` volume 中，普通容器重启不会丢失。

初始化 SQL：`docker/mysql/init/001-game-schema.sql`。初始化脚本只在空数据卷首次创建时执行。

## 验证

```bash
docker compose logs -f gameserver
nc -vz 127.0.0.1 19060
```
