FROM golang:1.21-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/server /app/server

RUN mkdir -p /app/data

ENV TZ=Asia/Shanghai
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8081
ENV SQLITE_PATH=/app/data/ordermonitor.db
ENV MONITOR_DATA_RETENTION_HOURS=12
ENV MONITOR_CLEANUP_INTERVAL_MINUTES=60
ENV SQLITE_VACUUM_ENABLED=true
ENV SQLITE_VACUUM_HOUR=4
ENV SQLITE_VACUUM_MINUTE=0
ENV REDIS_ENABLED=false
ENV GIN_MODE=release

EXPOSE 8081

CMD ["/app/server"]
