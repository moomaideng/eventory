# syntax=docker/dockerfile:1

FROM golang:1.26.5-alpine AS base

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

FROM base AS builder

ARG SERVICE=account

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./services/${SERVICE}/cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./services/${SERVICE}/cmd/migrate && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/seed ./services/${SERVICE}/cmd/seed

FROM alpine:3.22

RUN apk add --no-cache ca-certificates wget && \
    addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=builder /out/api /out/migrate /out/seed ./

USER app

HEALTHCHECK --interval=10s --timeout=3s --retries=5 --start-period=20s \
    CMD wget --no-verbose --tries=1 --output-document=/dev/null http://127.0.0.1:${HTTP_PORT}/health || exit 1

CMD ["./api"]
