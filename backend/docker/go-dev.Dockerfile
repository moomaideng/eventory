# syntax=docker/dockerfile:1
# Pinned Go hot-reload image for all backend services.

FROM golang:1.26.5-alpine

ARG AIR_VERSION=1.67.3

RUN apk add --no-cache wget && \
    go install github.com/air-verse/air@v${AIR_VERSION}

WORKDIR /app

COPY docker/go-dev-entrypoint.sh /usr/local/bin/go-dev-entrypoint.sh
RUN chmod +x /usr/local/bin/go-dev-entrypoint.sh

ENTRYPOINT ["go-dev-entrypoint.sh"]
