FROM golang:1.24-alpine AS builder
WORKDIR /srv
COPY pkg/ ./pkg/
COPY services/user/ ./services/user/
RUN cd services/user && go mod download && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /srv/user-service .

FROM alpine:3.20
WORKDIR /srv
COPY --from=builder /srv/user-service .
RUN apk add --no-cache curl && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
USER appuser:appgroup
EXPOSE 9091
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:9091/v1/health || exit 1
CMD ["./user-service"]
