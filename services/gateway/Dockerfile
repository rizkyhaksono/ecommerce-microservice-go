FROM golang:1.27.4-alpine AS builder
WORKDIR /srv
COPY pkg/ ./pkg/
COPY services/gateway/ ./services/gateway/
RUN cd services/gateway && go mod download && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /srv/gateway .

FROM alpine:3.24.4
RUN apk add --no-cache curl && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
WORKDIR /srv
COPY --from=builder --chown=appuser:appgroup /srv/gateway .
USER appuser:appgroup
EXPOSE 9090
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:9090/v1/health || exit 1
CMD ["./gateway"]
