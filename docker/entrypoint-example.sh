#!/bin/bash

set -e 

sudo cron -f &
sleep 1
sudo ps | grep cron 1>/dev/null
echo "[+] cron started"

echo "[+] Running as user `id`"

echo "[+] Starting pspy now..."
pspy 2>/dev/null
