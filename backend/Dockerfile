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
