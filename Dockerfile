FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init umami-proxy && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o proxy .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/proxy .

EXPOSE 8080
CMD ["./proxy"]