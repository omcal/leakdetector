# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies (if any)
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o leakcheck ./cmd/leakcheck

# Final stage - minimal image
FROM alpine:3.19

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/leakcheck /usr/local/bin/leakcheck

# Create a directory for mounting source code to analyze
RUN mkdir /src

# Set the working directory to /src for convenience
WORKDIR /src

# Default to JSON output for programmatic use
ENTRYPOINT ["leakcheck", "--json"]

# By default, scan current directory
CMD ["."]
