# Multi-stage build for Go Resource Monitor Application

# Stage 1: Build the application
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod ./

# Download dependencies (this layer will be cached if go.mod doesn't change)
RUN go mod download

# Copy source code
COPY main.go ./

# Build the application
# CGO_ENABLED=0 for a static binary
# -ldflags="-w -s" to strip debug information and reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o resource-monitor .

# Stage 2: Create the runtime image
FROM alpine:latest

# Add ca-certificates in case we need to make HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /build/resource-monitor /app/resource-monitor

# Set environment variables for Go runtime
ENV GOMAXPROCS=0
ENV GOGC=100

# Run the application
CMD ["/app/resource-monitor"]
