FROM golang:1.23 AS builder
WORKDIR /app
# Copy go mod and sum
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download
# Copy source code
COPY *.go ./
# Build
RUN go build -o main .

FROM alpine:latest  
# Set up certificates and other required tools
RUN apk --no-cache add ca-certificates
WORKDIR /root/
# Copy binary from previous stage
COPY --from=builder /app/main .
# Run binary
CMD ["./main"]