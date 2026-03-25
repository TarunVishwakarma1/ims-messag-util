# ==========================================
# STAGE 1: Builder
# ==========================================
FROM golang:1.25-alpine AS builder

# 1. Install CA certificates
RUN apk --no-cache add ca-certificates tzdata

# 2. Set working directory
WORKDIR /app

# 3. Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# 4. Copy the rest of the application code
COPY . .

# 5. Compile the statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /message-api ./main.go

# ==========================================
# STAGE 2: Production
# ==========================================
FROM scratch

# 1. Copy SSL certificates and timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 2. Copy the binary
COPY --from=builder /message-api /message-api

# 3. Copy templates directory
COPY --from=builder /app/templates /templates

# Expose port (Internal or host mapping)
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/message-api"]
