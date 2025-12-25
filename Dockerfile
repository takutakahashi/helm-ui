# Stage 1: Build frontend
FROM node:24-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache git
COPY api/go.mod api/go.sum ./api/
WORKDIR /app/api
RUN go mod download
COPY api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Stage 3: Final image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=backend-builder /server /app/server
COPY --from=frontend-builder /app/web/dist /app/static

ENV PORT=8080
EXPOSE 8080

CMD ["/app/server"]
