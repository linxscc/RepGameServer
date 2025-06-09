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

## 6. 上传至EC2

编辑文件：
```
在安全中设置仅当前用户可以完全控制
```

```
#ec2初次部署
#ec2连接
ssh -i RepGameKey.pem ubuntu@13.237.148.137

sudo apt-get update
sudo apt-get upgrade -y

# 安装必要的依赖
sudo apt-get install -y ca-certificates curl gnupg

# 添加 Docker 官方 GPG 密钥
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# 添加 Docker APT 源
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 更新包索引
sudo apt-get update

# 安装 Docker Engine
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

#DNS分配网站
namecheap
ZHOUKern


```
## RepGame 部署指南
#ec2连接
ssh -i RepGameKey.pem ubuntu@13.237.148.137  
#EC2拉取docker
#第一次
docker login

拉取镜像
sudo docker pull kernzs/repgame:latest

#第二次以后，单个服务启动
sudo docker stop repgame
sudo docker rm repgame
sudo docker run -d -p 80:80 -p 8000:8000 -p 9060:9060 -p 3306:3306 --name repgame kernzs/repgame:latest

#上传docker-compose.yml，第二次以后，整体服务执行
scp -i .\RepGameKey.pem .\docker-compose.yml ubuntu@13.237.148.137:/home/ubuntu/
ssh -i RepGameKey.pem ubuntu@13.237.148.137
cd /home/ubuntu

sudo docker compose down
sudo docker compose up -d
```

#查看Docker容器的日志
镜像日志
ssh -i "RepGameKey.pem" ubuntu@13.237.148.137 "sudo docker logs -f repgame"
实时查看
sudo docker logs -f repgame_allinone

zspersonaldomain.it.com

---


#RDS数据库
转发本地通道
ssh -i "RepGameKey.pem" -L 13306:repgame-database-0.cx2omeoogidr.ap-southeast-2.rds.amazonaws.com:3306 ubuntu@13.237.148.137
登录设置
127.0.0.1
13306
repgameadmin
repgameadmin

repgame-database-0.cx2omeoogidr.ap-southeast-2.rds.amazonaws.com
3306
repgameadmin
repgameadmin



> **注意：**
> - 前端生产环境建议用 `npm run build` 后，将 `build` 目录内容复制到 `nginx/html` 目录。
> - Nginx 启动后，访问 http://localhost/ 即可访问前端，API 请求会自动代理到后端。
> - nginx.conf 的 server_name 字段和 hosts 文件中的域名必须一致。
> - 修改 hosts 文件需管理员权限。
> - 线上服务器请将域名解析到公网 IP。
