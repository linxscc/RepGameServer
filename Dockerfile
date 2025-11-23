# 前端构建阶段 - 使用新的Vite TypeScript项目
FROM node:20 AS frontend-build

# 添加构建参数，确保每次构建时都有不同的缓存层
ARG BUILD_TIME
ENV BUILD_TIME=${BUILD_TIME}

WORKDIR /app/front

# 首先复制package.json相关文件，利用Docker缓存
COPY Front/myrepapp-vite/package.json Front/myrepapp-vite/package-lock.json* ./myrepapp-vite/
WORKDIR /app/front/myrepapp-vite
RUN npm install

# 然后复制源代码和资源文件
COPY Front/myrepapp-vite ./
# 添加构建时间戳，强制重新构建静态资源
ARG BUILD_DATE
ENV REACT_APP_BUILD_DATE=${BUILD_DATE}
RUN npm run build

# 后端构建阶段
FROM golang:1.24.3 AS backend-build
WORKDIR /app/goserver
COPY GoServer/go.mod GoServer/go.sum ./
RUN go mod download
COPY GoServer ./
RUN go build -o server main.go

# 生产镜像
FROM nginx:1.25
WORKDIR /app
# 拷贝前端静态资源到 nginx html 目录 - Vite构建输出目录为dist
COPY --from=frontend-build /app/front/myrepapp-vite/dist /usr/share/nginx/html
# 拷贝 nginx 配置
COPY nginx/conf/nginx.conf /etc/nginx/nginx.conf
# 拷贝后端可执行文件
COPY --from=backend-build /app/goserver/server /app/server

# 拷贝后端配置文件（重要：GoFrame需要读取配置来设置端口）
COPY --from=backend-build /app/goserver/manifest /app/manifest

# 设置 Docker 环境变量，使用 RDS 数据库
ENV DOCKER_BUILD=1

# 拷贝后端配置文件
COPY --from=backend-build /app/goserver/tcpgameserver/config /app/goserver/tcpgameserver/config
# 确保数据库服务目录存在
RUN mkdir -p /app/goserver/tcpgameserver/service
COPY --from=backend-build /app/goserver/tcpgameserver/service /app/goserver/tcpgameserver/service
# 启动脚本
COPY start.sh /app/start.sh
RUN chmod +x /app/start.sh

EXPOSE 80
EXPOSE 8000
EXPOSE 9060

CMD ["/app/start.sh"]