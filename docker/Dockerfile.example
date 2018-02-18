FROM debian:stretch

RUN apt-get update && apt-get -y install cron python3 sudo procps

# install root cronjob
COPY docker/var/spool/cron/crontabs /var/spool/cron/crontabs
RUN chmod 600 /var/spool/cron/crontabs/root
COPY docker/root/scripts /root/scripts

# install pspy
COPY bin/pspy64 /usr/bin/pspy

# set up unpriviledged user
# allows passwordless sudo to start cron as root on startup
RUN useradd -ms /bin/bash myuser && \
    adduser myuser sudo && \
    echo 'myuser ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
USER myuser

# deploy startup script
COPY docker/entrypoint-example.sh /entrypoint.sh
RUN sudo chmod +x /entrypoint.sh
CMD ["/entrypoint.sh"]


