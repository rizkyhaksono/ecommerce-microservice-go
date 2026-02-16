FROM golang:1.24-alpine AS builder
WORKDIR /srv
COPY services/gateway/ ./services/gateway/
RUN cd services/gateway && go mod download && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /srv/gateway .

FROM alpine:3.20
WORKDIR /srv
COPY --from=builder /srv/gateway .
RUN apk add --no-cache curl && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
USER appuser:appgroup
EXPOSE 9090
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:9090/v1/health || exit 1
CMD ["./gateway"]
