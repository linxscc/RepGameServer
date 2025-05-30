# 前端构建阶段
FROM node:20 AS frontend-build
WORKDIR /app/front
COPY Front/myrepapp ./myrepapp
WORKDIR /app/front/myrepapp
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
# 拷贝前端静态资源到 nginx html 目录
COPY --from=frontend-build /app/front/myrepapp/build /usr/share/nginx/html
# 拷贝 nginx 配置
COPY nginx/conf/nginx.conf /etc/nginx/nginx.conf
# 拷贝后端可执行文件
COPY --from=backend-build /app/goserver/server /app/server
# 拷贝后端配置文件
COPY --from=backend-build /app/goserver/tcpgameserver/config/response_codes.json /app/goserver/tcpgameserver/config/response_codes.json
# 启动脚本
COPY start.sh /app/start.sh
RUN chmod +x /app/start.sh

EXPOSE 80
EXPOSE 8000
EXPOSE 9060

CMD ["/app/start.sh"]