#!/bin/sh
# Start Go backend
/app/server &

# Enable HTTPS config only when cert files are mounted.
if [ -f /etc/letsencrypt/live/zsdimain.site/fullchain.pem ] && [ -f /etc/letsencrypt/live/zsdimain.site/privkey.pem ]; then
  echo "[start] HTTPS cert detected, enabling TLS config"
  cp /etc/nginx/nginx.https.conf /etc/nginx/nginx.conf
else
  echo "[start] HTTPS cert not found, using HTTP-only config"
  cp /etc/nginx/nginx.http.conf /etc/nginx/nginx.conf
fi

nginx -t || exit 1

# Pick up certificates renewed by the Certbot sidecar without stopping the app.
(while sleep 21600; do nginx -s reload || true; done) &

nginx -g "daemon off;"
