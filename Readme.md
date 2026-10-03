# RepGame TCP Server

`RepServer` 是专门维护游戏后端服务器的分支。它只包含：

- Go TCP 游戏服务器：`GoServer/tcpgameserver`
- 独立游戏服务器入口：`GoServer/main.go`
- MySQL 8.4 与 RepGame 初始化结构
- 可选的 Namecheap DDNS 更新服务
- 未来游戏后台管理界面的空目录：`Front/game-admin`

原 Web 网站、商城前后端、Nginx、HTTPS 和 Web API 均不包含在此分支。

## 分支策略

`RepServer` 是长期独立维护和部署的游戏服务器分支。游戏服务端的修改直接提交并推送到 `RepServer`，不需要创建 PR，也不合并进入 `main`。`main` 继续用于原网站项目，两条分支各自构建、部署和维护。

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

## GitHub Actions 自动部署

`.github/workflows/deploy-repserver.yml` 会在 `RepServer` 分支收到新提交后，在 Mac mini 上自动完成：

1. 拉取触发工作流的最新提交。
2. 校验本机 Docker 和部署环境文件。
3. 执行 `docker compose up -d --build`。
4. 检查游戏容器和 TCP 端口；失败时输出最近日志。

工作流使用 Mac mini 本地的 `.env`，不会把数据库或 DDNS 密钥上传到 GitHub。默认路径为：

```text
/Users/shuangzhou/Documents/ChatGPT/RepGameServer/.env
```

如路径不同，在 GitHub 仓库的 **Settings → Secrets and variables → Actions → Variables** 中创建变量 `REPGAME_ENV_FILE`。

### Mac mini 首次配置

进入 GitHub 仓库的 **Settings → Actions → Runners → New self-hosted runner**，选择 macOS 和这台 Mac 对应的架构，按照页面提供的命令下载并注册 Runner。注册完成后，在 Runner 目录执行：

```bash
./svc.sh install
./svc.sh start
./svc.sh status
```

Runner 必须使用能够访问 Docker Desktop 的 macOS 用户运行。Docker Desktop 需设置为登录时自动启动；游戏容器自身使用 `restart: unless-stopped`，Docker 恢复后会自动重新启动。
