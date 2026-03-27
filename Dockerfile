FROM node:20 AS frontend-build

ARG BUILD_TIME
ENV BUILD_TIME=${BUILD_TIME}

WORKDIR /app/front
COPY Front/myrepapp-vite/package.json Front/myrepapp-vite/package-lock.json* ./myrepapp-vite/
WORKDIR /app/front/myrepapp-vite
RUN npm install

COPY Front/myrepapp-vite ./
ARG BUILD_DATE
ENV REACT_APP_BUILD_DATE=${BUILD_DATE}
RUN npm run build

FROM golang:1.24.3 AS backend-build
WORKDIR /app/goserver
COPY GoServer/go.mod GoServer/go.sum ./
RUN go mod download
COPY GoServer ./
RUN go build -o server main.go

FROM nginx:1.25
WORKDIR /app

COPY --from=frontend-build /app/front/myrepapp-vite/dist /usr/share/nginx/html
COPY nginx/conf/nginx.http.conf /etc/nginx/nginx.http.conf
COPY nginx/conf/nginx.https.conf /etc/nginx/nginx.https.conf
COPY nginx/conf/app-locations.conf /etc/nginx/app-locations.conf

COPY --from=backend-build /app/goserver/server /app/server
COPY --from=backend-build /app/goserver/manifest /app/manifest

ENV DOCKER_BUILD=1

COPY --from=backend-build /app/goserver/tcpgameserver/config /app/goserver/tcpgameserver/config
RUN mkdir -p /app/goserver/tcpgameserver/service
COPY --from=backend-build /app/goserver/tcpgameserver/service /app/goserver/tcpgameserver/service

COPY start.sh /app/start.sh
RUN chmod +x /app/start.sh

EXPOSE 80
EXPOSE 443
EXPOSE 8000
EXPOSE 9060

CMD ["/app/start.sh"]
