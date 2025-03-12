# Build stage
FROM golang:1.23.7-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the applications
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate

# Final stage
FROM gcr.io/distroless/static

# Set working directory
WORKDIR /app

# Copy binaries from build stage
COPY --from=builder /app/api .
COPY --from=builder /app/migrate .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Run the application
CMD ["./api"]
