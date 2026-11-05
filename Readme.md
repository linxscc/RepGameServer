# RepGame TCP Server

`RepServer` 是专门维护游戏后端服务器的分支。它只包含：

- Go TCP 游戏服务器：`GoServer/tcpgameserver`
- 独立游戏服务器入口：`GoServer/main.go`
- MySQL 8.4 与 RepGame 初始化结构
- 可选的 Namecheap DDNS 更新服务
- 未来游戏后台管理界面的空目录：`Front/game-admin`

原 Web 网站、商城前后端、Nginx、HTTPS 和 Web API 均不包含在此分支。

服务通过 Docker Compose 运行，推送到 `RepServer` 分支后由 GitHub Actions 自动部署。部署与运维说明维护在本地文档中，不纳入版本库。
