#!/bin/bash

### test covereage

if [ -z ${CC_TEST_REPORTER_ID+x} ]; then
  echo "[+] skipping test coverage"
else
  echo "[+] reporting test coverage"
  curl -L https://codeclimate.com/downloads/test-reporter/test-reporter-latest-linux-amd64 > ./cc-test-reporter
  chmod +x ./cc-test-reporter

  # git config --global user.email "dummy@example.com"
  # git config --global user.name "Mr Robot"
  # git init
  # git add .
  # git commit -m 'commit that makes cc test reporter happy'

  ./cc-test-reporter before-build

  for pkg in $(go list ./... | grep -v main); do
    go test -coverprofile=$(echo $pkg | tr / -).cover $pkg
  done
  echo "mode: set" > c.out
  grep -h -v "^mode:" ./*.cover >> c.out
  rm -f *.cover

  ./cc-test-reporter after-build

  rm c.out

  rm ./cc-test-reporter
fi

### integration test

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
