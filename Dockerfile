# Stage 1: Build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Stage 2: Run
FROM alpine:latest
RUN adduser -D -g '' appuser
COPY --from=builder /app/server /app/server
EXPOSE ${PORT:-8080}
USER appuser
ENTRYPOINT ["/app/server"]
