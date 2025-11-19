# Используем многостадийную сборку
FROM golang:1.24-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы модулей для кэширования слоев
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download && go mod verify

# Копируем исходный код
COPY . .

# Собираем приложение с оптимизацией
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o log-metrics-simulator .

# Финальный образ
FROM alpine:3.19

# Устанавливаем зависимости для runtime
RUN apk --no-cache add ca-certificates tzdata curl

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S app && \
    adduser -u 1001 -S app -G app

# Создаем рабочие директории с правильными разрешениями
RUN mkdir -p /app/logs /app/data && \
    chown -R app:app /app && \
    chmod -R 755 /app

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder --chown=app:app /app/log-metrics-simulator .

# Копируем конфигурационные файлы
COPY --chown=app:app config.yaml ./

# Переключаемся на непривилегированного пользователя
USER app

# Экспонируем порт
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Метаданные
LABEL maintainer="log-metrics-simulator" \
      version="1.0.0" \
      description="Log and Metrics Simulator for E-commerce"

# Запускаем приложение
CMD ["./log-metrics-simulator"]