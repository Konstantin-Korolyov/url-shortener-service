# ---- Stage 1: build ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем статически скомпилированный бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o url-shortener ./cmd/server

# ---- Stage 2: final image ----
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/url-shortener .

# Открываем порт
EXPOSE 8080

# Запускаем
CMD ["./url-shortener"]