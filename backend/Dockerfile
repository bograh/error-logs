# # ---------- Build stage ----------
# FROM golang:1.24-alpine AS builder

# WORKDIR /app

# # Cache deps first
# COPY go.mod go.sum ./
# RUN go mod download

# # Copy source and build
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# # ---------- Runtime stage ----------
# FROM alpine:latest AS runner

# RUN apk --no-cache add ca-certificates

# WORKDIR /root/

# # Copy the compiled binary
# COPY --from=builder /app/main .

# # Run the binary (NOT `go`)
# CMD ["./main"]


FROM golang:1.24-alpine

WORKDIR /app

# Install git (needed for Go modules sometimes)
RUN apk add --no-cache git

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the backend source
COPY . .

# Default command (you can override in docker-compose)
CMD ["go", "run", "main.go"]
