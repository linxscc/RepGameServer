# 前端构建阶段 - 使用新的Vite TypeScript项目
FROM node:20 AS frontend-build
WORKDIR /app/front
COPY Front/myrepapp-vite ./myrepapp-vite
WORKDIR /app/front/myrepapp-vite
RUN npm install && npm run build

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