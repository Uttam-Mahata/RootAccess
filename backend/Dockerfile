FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -tags production -o main ./cmd/api/main.go
# Build the admin tool
RUN CGO_ENABLED=0 GOOS=linux go build -o admin-tool ./cmd/admin/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/main .
COPY --from=builder /app/admin-tool .

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
