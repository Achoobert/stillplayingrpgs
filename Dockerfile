# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies for cgo (needed for sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go-backend/go.mod go-backend/go.sum ./
RUN go mod download

COPY go-backend/ .
# Enable CGO for sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/server

# Final stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/main .
COPY go-backend/templates ./templates
COPY go-backend/static ./static

EXPOSE 30011
CMD ["./main"]
