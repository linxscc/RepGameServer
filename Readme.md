# Start and close commands

## 1. Start the GoFrame backend

```powershell
cd GoServer
go run main.go
go build -o server main.go
```

Close:
- Simply press `Ctrl+C` in the run window.

---

## 2. Start the React frontend

```powershell
cd Front\myrepapp
npm start
```
Package
npm run build

Close:
- Simply press `Ctrl+C` in the run window.

---

## 3. Start Nginx

```powershell
cd nginx
./nginx.exe -p . -c conf/nginx.conf
```

Shutdown:
```powershell
./nginx.exe -s stop
```

---

## 4. Docker integrated deployment

```powershell
# Build and start all services (frontend + backend + Nginx)
docker compose up --build
```

Shutdown:
```powershell
docker compose down
```
Rebuild and start
Build and start all services
docker compose build --no-cache
Or just start (no rebuild)
docker compose up

Restart container
docker compose down
docker compose up --build

#Local upload docker
#No cache build
docker build --no-cache -t kernzs/repgame:latest .
#Cache build
docker build -t kernzs/repgame:latest .
docker push kernzs/repgame:latest

---

## 5. Local hosts file configuration (make sure the domain name is consistent with nginx.conf)

Edit the file:
```
C:\Windows\System32\drivers\etc\hosts
```
Add a line (consistent with the server_name of nginx.conf):
```
127.0.0.1 zspersonaldomain.com
```

This way, accessing http://zspersonaldomain.com/ in the browser will point to the local Nginx.

---
