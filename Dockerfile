FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -o /bin/post ./cmd/post/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl
COPY --from=builder /bin/post /usr/local/bin/post
ENV PORT="8830" DATA_DIR="/data" RETENTION_DAYS="30" POST_LICENSE_KEY=""
EXPOSE 8830
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -sf http://localhost:8830/health || exit 1
ENTRYPOINT ["post"]
