#!/bin/bash

sudo cron -f &
sleep 1
sudo ps | grep cron 1>/dev/null
echo "[+] cron started"

echo "[+] Running as user `id`"

echo "[+] Executing test"
# exec /bin/bash
rm /home/myuser/log.txt
bin/pspy > /home/myuser/log.txt &

for i in `seq 1 25`; do
    echo "Waiting for cron job detection..."
    sleep 5;

    QUERY_RESULT=$(grep ' | passwd myuser' /home/myuser/log.txt | grep -v grep)
    if [ "'$QUERY_RESULT'" != "''"  ]; then
        echo "Cron job execution detected!"
        echo "Complete log of pspy (may contain commands run in this test):"
        cat /home/myuser/log.txt
        exit 0
    fi
done
echo "Failed to detect cron job..."
exit 1
