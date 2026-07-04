# 后台 Job 服务构建
FROM golang:1.26.4-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/job ./cmd/job

# 运行阶段
FROM alpine:3.22

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata \
    && adduser -D -s /bin/sh appuser

COPY --from=builder /app/job .
COPY --from=builder /app/configs ./configs

USER appuser

ENTRYPOINT ["./job"]
