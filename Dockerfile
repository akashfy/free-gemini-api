# Step 1: Build the Go binary
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

# Step 2: Final minimal run image
FROM alpine:latest
RUN apk add --no-cache ffmpeg ca-certificates
WORKDIR /app
COPY --from=builder /app/main /app/main
COPY --from=builder /app/assets /app/assets

WORKDIR /app/data
EXPOSE 8001 9226
CMD ["/app/main"]
