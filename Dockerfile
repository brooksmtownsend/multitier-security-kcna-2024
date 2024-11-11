# Stage 1: Build the Go binary
FROM golang:1.23 AS builder

# Set the working directory
WORKDIR /app

# Copy the current directory contents into the container
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o hello .

# Stage 2: Create a minimal scratch container with the binary
FROM gcr.io/distroless/static-debian12

# Copy the binary from the builder stage
COPY --from=builder /app/hello /hello

# Set the entrypoint to the binary
ENTRYPOINT ["/hello"]
