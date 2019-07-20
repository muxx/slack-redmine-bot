# build binary

FROM golang:alpine AS builder

RUN apk --update --no-cache add --virtual build-dependencies \
    git make \
    && git clone https://github.com/muxx/slack-redmine-bot /root/repo \
    && make build -C /root/repo \
    && cp /root/repo/bin/slack-redmine-bot / \
    && rm -rf /root/repo \
    && apk update && apk del build-dependencies

# build image

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /slack-redmine-bot /slack-redmine-bot

WORKDIR /

ENTRYPOINT ["/slack-redmine-bot", "--config", "/config.yml"]
