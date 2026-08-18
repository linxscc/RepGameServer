#!/bin/sh
set -eu

domain="${CERTBOT_DOMAIN:-zsdimain.site}"
certificate="/etc/letsencrypt/live/${domain}/fullchain.pem"

while [ ! -f "$certificate" ]; do
    echo "[certbot] requesting certificate for ${domain} and www.${domain}"
    if certbot certonly \
        --webroot \
        --webroot-path /var/www/certbot \
        --domain "$domain" \
        --domain "www.$domain" \
        --agree-tos \
        --non-interactive \
        --register-unsafely-without-email \
        --preferred-challenges http; then
        echo "[certbot] certificate issued successfully"
        break
    fi
    echo "[certbot] issuance failed; retrying in 5 minutes"
    sleep 300
done

while true; do
    certbot renew --webroot --webroot-path /var/www/certbot --quiet || true
    sleep 43200
done
