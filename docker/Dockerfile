FROM golang:1.9-stretch

RUN apt-get update && apt-get -y install cron sudo
COPY docker/etc/cron.d /etc/cron.d
COPY docker/scripts /scripts

RUN useradd -ms /bin/bash myuser && \
    adduser myuser sudo && \
    echo 'myuser ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
USER myuser


WORKDIR /go/src/github.com/dominicbreuker


