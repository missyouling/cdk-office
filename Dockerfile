# CDK-Office Dockerfile

# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o cdk-office cmd/server/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/cdk-office .

# Copy configuration files
COPY --from=builder /app/config ./config

# Expose port
EXPOSE 8080

# Command to run the application
CMD ["./cdk-office"]