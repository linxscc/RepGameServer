# 启动与关闭命令

## 1. 启动 GoFrame 后端

```powershell
cd GoServer
go run main.go
go build -o server main.go
```

关闭：
- 直接在运行窗口按 `Ctrl+C` 即可。

---

## 2. 启动 React 前端

```powershell
cd Front\myrepapp
npm start
```
打包
npm run build

关闭：
- 直接在运行窗口按 `Ctrl+C` 即可。

---

## 3. 启动 Nginx

```powershell
cd nginx
./nginx.exe -p . -c conf/nginx.conf
```

关闭：
```powershell
./nginx.exe -s stop
```

---

## 4. Docker 一体化部署

```powershell
# 构建并启动所有服务（前端+后端+Nginx）
docker compose up --build
```

关闭：
```powershell
docker compose down
```
重新构建并启动
构建并启动所有服务
docker compose build --no-cache
或仅启动（不重新构建）
docker compose up

重启容器
docker compose down
docker compose up --build

#本地上传docker
#无缓存构建
docker build --no-cache -t kernzs/repgame:latest .
#缓存构建
docker build -t kernzs/repgame:latest .
docker push kernzs/repgame:latest

---

## 5. 本地 hosts 文件配置（确保域名和 nginx.conf 一致）

编辑文件：
```
C:\Windows\System32\drivers\etc\hosts
```
添加一行（与 nginx.conf 的 server_name 保持一致）：
```
127.0.0.1   zspersonaldomain.com
```

这样在浏览器访问 http://zspersonaldomain.com/ 就会指向本地 Nginx。

---


