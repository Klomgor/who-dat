# Multi-stage Dockerfile for Who-Dat
# Produces a minimal ~15MB Alpine-based image

# Stage 1: Build the frontend
FROM node:20-alpine as frontend-builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY src ./src
COPY public ./public
COPY tsconfig.json vite.config.js ./
RUN npm run build

# Stage 2: Build the Go application
FROM golang:1.21-alpine as go-builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/dist ./cmd/server/dist
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o who-dat \
    ./cmd/server

# Stage 3: Final minimal image
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=go-builder /build/who-dat .
RUN chown -R appuser:appuser /app

USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ping || exit 1

ENTRYPOINT ["./who-dat"]
