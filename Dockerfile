# Stage 1: Frontend build
FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Go build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dist ./frontend/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# Stage 3: Run
FROM alpine:latest
RUN adduser -D -g '' appuser
COPY --from=builder /app/server /app/server
EXPOSE ${PORT:-8080}
USER appuser
ENTRYPOINT ["/app/server"]
