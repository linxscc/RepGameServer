#!/bin/sh
set -eu

if [ -z "${DDNS_PASSWORD:-}" ]; then
    echo "[ddns] NAMECHEAP_DDNS_PASSWORD is empty; refusing to start"
    exit 1
fi

domain="${DDNS_DOMAIN:-zsdimain.site}"
host="${DDNS_HOST:-@}"
interval="${DDNS_INTERVAL:-300}"
last_ip=""

echo "[ddns] monitoring ${host}.${domain} every ${interval}s"

while true; do
    current_ip="$(curl -fsS --max-time 15 https://dynamicdns.park-your-domain.com/getip || true)"

    if [ -z "$current_ip" ]; then
        echo "[ddns] unable to determine the public IPv4 address"
    elif [ "$current_ip" != "$last_ip" ]; then
        response="$(curl -fsS --max-time 20 --get \
            --data-urlencode "host=$host" \
            --data-urlencode "domain=$domain" \
            --data-urlencode "password=$DDNS_PASSWORD" \
            --data-urlencode "ip=$current_ip" \
            https://dynamicdns.park-your-domain.com/update || true)"

        if echo "$response" | grep -q '<ErrCount>0</ErrCount>'; then
            echo "[ddns] ${host}.${domain} updated to ${current_ip}"
            last_ip="$current_ip"
        else
            echo "[ddns] update failed; verify the A+Dynamic DNS record and DDNS password"
        fi
    fi

    sleep "$interval"
done
