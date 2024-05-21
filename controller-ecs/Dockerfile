# Use the official Golang image to create a build artifact.
FROM golang:1.22 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN GOOS=linux go build -o main ./cmd/main.go

# Start a new stage from scratch
FROM alpine:latest

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

COPY ./prometheus /ecs/shared-config

# Expose port 8080 to the outside world
EXPOSE 80

# Command to run the executable
CMD ["./main"]