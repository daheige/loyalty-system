# API 服务构建
FROM golang:1.26.4-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/api

# 运行阶段
FROM alpine:3.22

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata wget \
    && adduser -D -s /bin/sh appuser

COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./server"]
