# RepGameServer 本机部署

项目包含：Vite/React Web 前端、Go HTTP API（`8000`）、Go TCP 游戏服务 `tcpgameserver`（`9060`）、Nginx Web 入口（`80`/`443`）和 MySQL 8.4（仅绑定本机 `127.0.0.1:13306`）。

## 一键启动

要求安装 Docker Desktop（或 Docker Engine + Compose）。

```bash
cp .env.example .env
```

打开 `.env`，至少替换 `DB_PASSWORD`、`MYSQL_ROOT_PASSWORD` 和 `VOYARA_JWT_SECRET`（至少 32 个随机字符），然后运行：

```bash
docker compose up -d --build
docker compose ps
```

访问地址：

- Web：<http://localhost>
- HTTP API：<http://localhost:8000>
- Swagger：<http://localhost:8000/swagger>
- TCP 游戏服务：`localhost:9060`
- MySQL：`127.0.0.1:13306`

查看日志或停止：

```bash
docker compose logs -f app
docker compose down
```

数据库、商品图片保存在 Docker volumes 中。普通 `docker compose down` 不会删除数据；只有明确执行 `docker compose down -v` 才会清空。

本地下载文件放入项目的 `downloads/` 目录，可通过 `/download/文件名` 或 `/downloads/文件名` 访问。

## 局域网访问

Compose 默认监听 `0.0.0.0`。查询本机局域网 IP 后，同一网络设备可访问：

```text
http://你的局域网IP
你的局域网IP:9060
```

请在 macOS/Windows 防火墙中允许 Docker 接收 `80/tcp` 和 `9060/tcp`。`8000` 是调试用直连 API；公开访问时建议只开放 Web `80/443` 和游戏 `9060`。

## 外网访问

### 方案 A：路由器端口转发

给本机设置固定局域网 IP，并在路由器配置 TCP 转发：

| 公网端口 | 本机端口 | 用途 |
|---:|---:|---|
| 80 | 80 | Web HTTP |
| 443 | 443 | Web HTTPS（配置证书后） |
| 9060 | 9060 | TCP 游戏服务 |

然后使用 `http://公网IP` 访问 Web，游戏客户端连接 `公网IP:9060`。不要转发 MySQL 的 `13306`。如果运营商使用 CGNAT、没有独立公网 IPv4，路由器转发不会生效，需要采用内网穿透或云隧道。

### 方案 B：域名与 HTTPS

域名指向本机并开放 TCP 80/443 后，启动内置 Certbot 服务：

```bash
docker compose --profile tls up -d certbot
docker compose logs -f certbot
docker compose restart app
```

证书保存在 `letsencrypt_data` Docker volume 中。Certbot 每 12 小时检查续期，Nginx 定期重新加载新证书。

若前端和 API 使用不同域名，在 `.env` 中填写逗号分隔的来源：

```text
ALLOWED_ORIGINS=https://example.com,https://www.example.com
```

云隧道通常只代理 HTTP/HTTPS；原生 TCP `9060` 还需要路由器转发、支持 TCP 的隧道或 VPN。

### 公网 IP 变化：Namecheap DDNS

SoftBank 重新拨号后公网 IPv4 可能变化。项目内置了 Namecheap DDNS 更新容器，每 5 分钟检查一次公网地址，仅在地址变化时更新根域名。

1. Namecheap → Domain List → Manage → Advanced DNS。
2. 将根域名 `@` 的记录类型改为 `A+ Dynamic DNS Record`；`www` 可继续使用指向 `zsdimain.site` 的 CNAME。
3. 在同一页面的 Dynamic DNS 区域启用功能并复制 Dynamic DNS Password。
4. 把该密码写入本机 `.env` 的 `NAMECHEAP_DDNS_PASSWORD`。不要填写 Namecheap 账户密码。
5. 启动一次 DDNS profile：

```bash
docker compose --profile ddns up -d ddns
docker compose logs -f ddns
```

容器使用 `restart: unless-stopped`，Docker 恢复运行后会自动继续更新。macOS 还应开启“断电后自动启动”，并设置 Docker Desktop 登录后自动启动。

## AWS 兼容选项

默认 `STORAGE_DRIVER=local`，商品图片保存在本机。若仍想使用 S3，将它改为 `s3`，并填写 `.env` 中的 `AWS_*` 参数。开发环境邮件验证码输出到应用日志，不依赖 AWS SES。

## 重新初始化数据库

初始化 SQL 只在 MySQL volume 第一次创建时运行。若修改了 `docker/mysql/init/001-local-schema.sql` 并希望从空库重建：

```bash
docker compose down -v
docker compose up -d --build
```

此操作会删除本地数据库和上传文件，请先备份。
