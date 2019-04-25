FROM golang:1.12-stretch

RUN apt-get update && apt-get -y install cron python3 sudo procps

### Test and build ###

# copy source code
COPY main.go /go/src/github.com/dominicbreuker/pspy/main.go
COPY cmd /go/src/github.com/dominicbreuker/pspy/cmd
COPY internal /go/src/github.com/dominicbreuker/pspy/internal
COPY vendor /go/src/github.com/dominicbreuker/pspy/vendor
COPY .git /go/src/github.com/dominicbreuker/pspy/.git

# run tests
WORKDIR /go/src/github.com/dominicbreuker/pspy
RUN go test ./...
# build executable
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o bin/pspy main.go

### Prepare integration test ###
# install root cronjob
COPY docker/var/spool/cron/crontabs /var/spool/cron/crontabs
RUN chmod 600 /var/spool/cron/crontabs/root
COPY docker/root/scripts /root/scripts

# set up unpriviledged user
# allows passwordless sudo to start cron as root on startup
RUN useradd -ms /bin/bash myuser && \
    adduser myuser sudo && \
    echo 'myuser ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
USER myuser
RUN sudo chown -R myuser:myuser /go/*

# drop into bash shell
COPY docker/entrypoint-testing.sh /entrypoint.sh
RUN sudo chmod +x /entrypoint.sh
CMD ["/entrypoint.sh"]
