#!/bin/sh
# 启动 GoFrame 后端
/app/server &
# 启动 Nginx
nginx -g "daemon off;"