# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies (if any)
RUN go mod download

# Copy source code
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 go build -o server main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server .

# Expose port
EXPOSE 8081

# Run the application
CMD ["./server"]
