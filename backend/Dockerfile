# syntax=docker/dockerfile:1

FROM golang:1.26.5-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.22

RUN apk add --no-cache ca-certificates wget && \
    addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=builder /out/api /out/migrate ./

USER app
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
    CMD wget --no-verbose --tries=1 --output-document=/dev/null http://127.0.0.1:8080/health || exit 1

CMD ["./api"]
