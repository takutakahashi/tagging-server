# Use the official Golang image as the base for building the binary
FROM golang:1.23-bullseye as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifest and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o tagging-server main.go

FROM litestream/litestream:latest as litestream
# Use a smaller base image for the final container
FROM debian:bullseye

# Install SQLite
RUN apt-get update && apt-get install -y ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/tagging-server /app/tagging-server
COPY --from=litestream /usr/local/bin/litestream /usr/local/bin/litestream

# Copy the entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Set the environment variables for the database path and Litestream restore URI (example values)
# These should be overridden at runtime.
ENV LITESTREAM_REPLICA_URI=""
ENV DB_PATH="/app/data/targets_tags.db"

# Expose the server port
EXPOSE 8080

# Set the entrypoint to execute the entrypoint.sh script
ENTRYPOINT ["/app/entrypoint.sh"]
