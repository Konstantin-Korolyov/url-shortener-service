Структура:

markdown
# URL Shortener Service

Production-ready сервис сокращения ссылок с аналитикой, построенный на Go, PostgreSQL, Redis и Kafka.

## Возможности
- Создание коротких ссылок через REST API.
- Редирект по короткой ссылке.
- Сбор статистики переходов (IP, user-agent, referer, timestamp).
- Асинхронная запись статистики через Kafka.
- Кеширование популярных ссылок в Redis.
- Контейнеризация с Docker Compose.
- Graceful shutdown.
- (Добавить позже: rate limiting, авторизация, метрики)

## Технологии
- **Go 1.25+** — основной язык.
- **PostgreSQL** — хранение ссылок и событий.
- **Redis** — кеширование.
- **Kafka** — очередь для асинхронной обработки кликов.
- **Docker / Docker Compose** — запуск всех сервисов.
- **pgx** — драйвер для PostgreSQL.
- **go-redis** — клиент Redis.
- **segmentio/kafka-go** — работа с Kafka.

## Архитектура
(Вставь схему или текстовое описание, например:)
1. Клиент → POST /shorten → приложение → вставка в PostgreSQL → ответ с короткой ссылкой.
2. Клиент → GET /r/{code} → приложение:
   - Проверка Redis (если есть → редирект + отправка события в Kafka)
   - Иначе запрос в PostgreSQL → кеширование в Redis → редирект + отправка события.
3. Фоновый consumer читает события из Kafka и сохраняет в таблицу click_events.

## Запуск проекта

### Требования
- Docker и Docker Compose установлены.

### Шаги
1. Клонировать репозиторий:
   ```bash
   git clone https://github.com/yourusername/url-shortener-service.git
   cd url-shortener-service


Запустить все сервисы:
docker-compose up -d
Приложение будет доступно на http://localhost:8080.


Остановка всех сервисов:
docker-compose down
API Endpoints
POST /shorten
Создаёт короткую ссылку.


Request body:
json
{
  "url": "https://example.com"
}

Response:
json
{
  "short_url": "http://localhost:8080/r/abc123"
}

GET /r/{code}
Редирект на оригинальный URL.

GET /health
Проверка работоспособности.

Примеры использования:
# Создать короткую ссылку
curl -X POST http://localhost:8080/shorten -H "Content-Type: application/json" -d '{"url":"https://google.com"}'

# Перейти по короткой ссылке (в браузере)
http://localhost:8080/r/abc123

Переменные окружения:
Все настройки задаются через переменные окружения (см. docker-compose.yml):
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_DB
KAFKA_BROKER, KAFKA_TOPIC
SERVER_PORT


Планы по улучшению:
Добавить rate limiting.
Добавить авторизацию по API-ключу.
Определение страны по IP.
Метрики Prometheus и дашборд Grafana.
Структурированное логирование.
Обработка коллизий коротких кодов.
<<<<<<< HEAD
CI/CD с GitHub Actions.


Лицензия
MIT (или другая)
=======
CI/CD с GitHub Actions.
>>>>>>> 130625a14b40d30273740439fb0e6806aab1be2b
