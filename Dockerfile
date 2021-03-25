FROM debian:buster-slim

RUN apt-get update && apt-get install -y \
    curl git
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["slackln"]
CMD [ "-h" ]

COPY slackln_*.deb /tmp/
RUN dpkg -i /tmp/slackln_*.deb
