#!/bin/sh
# Start Go backend
/app/server &

# Enable HTTPS config only when cert files are mounted.
if [ -f /etc/nginx/certs/fullchain.pem ] && [ -f /etc/nginx/certs/privkey.pem ]; then
  echo "[start] HTTPS cert detected, enabling TLS config"
  cp /etc/nginx/nginx.https.conf /etc/nginx/nginx.conf
else
  echo "[start] HTTPS cert not found, using HTTP-only config"
  cp /etc/nginx/nginx.http.conf /etc/nginx/nginx.conf
fi

nginx -t || exit 1
nginx -g "daemon off;"
