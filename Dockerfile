# Stage 1: Frontend build
FROM node:22-alpine AS frontend
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL
WORKDIR /app
RUN npm install -g npm@latest
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
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/migrate ./cmd/migrate/

# Stage 3: Run
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache postgresql17-client && \
    adduser -D -g '' appuser
COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /app/db/migrations /app/db/migrations
COPY prod/start-prod.sh /app/start-prod.sh
RUN chmod +x /app/start-prod.sh
EXPOSE ${PORT:-8080}
USER appuser
ENTRYPOINT ["/app/start-prod.sh"]
