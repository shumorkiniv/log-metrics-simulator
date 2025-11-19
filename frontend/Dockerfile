# Многостадийная сборка для React приложения
FROM node:20-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы package.json и package-lock.json для кэширования слоев
COPY package*.json ./

# Устанавливаем все зависимости (включая devDependencies для сборки)
RUN npm ci

# Копируем исходный код
COPY . .

# Собираем приложение
RUN npm run build

# Проверяем, что сборка прошла успешно
RUN ls -la /app/dist

# Продакшн стадия с nginx
FROM nginx:1.25-alpine AS production

# Устанавливаем зависимости
RUN apk add --no-cache curl

# Создаем пользователя nginx если его нет
RUN addgroup -g 101 -S nginx || true
RUN adduser -S -D -H -u 101 -h /var/cache/nginx -s /sbin/nologin -G nginx -g nginx nginx || true

# Создаем директории для nginx с правильными разрешениями
RUN mkdir -p /var/cache/nginx /var/log/nginx /etc/nginx/conf.d /run/nginx && \
    touch /var/run/nginx.pid && \
    chown -R nginx:nginx /var/cache/nginx /var/log/nginx /var/run/nginx.pid /run/nginx

# Удаляем дефолтную конфигурацию nginx
RUN rm -f /etc/nginx/conf.d/default.conf

# Копируем конфигурацию nginx
COPY nginx.conf /etc/nginx/nginx.conf

# Копируем собранное приложение
COPY --from=builder /app/dist /usr/share/nginx/html

# Устанавливаем правильные права на статические файлы
RUN chown -R nginx:nginx /usr/share/nginx/html

# Экспонируем порт
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/health || exit 1

# Метаданные
LABEL maintainer="log-metrics-simulator" \
      version="1.0.0" \
      description="Frontend for Log and Metrics Simulator"

# Запускаем nginx от пользователя nginx
USER nginx
CMD ["nginx", "-g", "daemon off;"]