Структура:

markdown
# URL Shortener Service

Production-ready сервис сокращения ссылок с аналитикой, построенный на **Go**, **PostgreSQL**, **Redis** и **Kafka**.

## Возможности
- Создание коротких ссылок через REST API.
- Редирект по короткой ссылке с автоматическим сбором статистики.
- Детальная аналитика переходов: IP, User-Agent, Referer, временная метка, страна (по IP).
- Асинхронная запись событий через Kafka (не блокирует ответ пользователя).
- Кеширование популярных ссылок в Redis для ускорения редиректов.
- Авторизация по API-ключу (привязка ссылок к пользователям).
- Rate limiting (ограничение запросов с одного IP) для защиты от злоупотреблений.
- Метрики Prometheus и дашборд Grafana для мониторинга.
- Полная контейнеризация с Docker Compose.
- Graceful shutdown.
- Структурированное логирование (slog).
- Обработка коллизий коротких кодов (повторные попытки при уникальности).

## Технологии
- **Go 1.21+** — основной язык.
- **PostgreSQL** — хранение ссылок, пользователей и событий.
- **Redis** — кеширование ссылок.
- **Kafka** — очередь для асинхронной обработки кликов.
- **Docker / Docker Compose** — запуск всех сервисов.
- **pgx** — драйвер для PostgreSQL.
- **go-redis** — клиент Redis.
- **segmentio/kafka-go** — работа с Kafka.
- **Prometheus + Grafana** — сбор и визуализация метрик.
- **GeoIP2 (MaxMind)** — определение страны по IP.

## Архитектура

1. Клиент - `POST /shorten` - приложение - проверка API-ключа - вставка в PostgreSQL - ответ с короткой ссылкой.
2. Клиент - `GET /r/{code}` - приложение:
   - Проверка Redis (если есть - редирект + отправка события в Kafka)
   - Иначе запрос в PostgreSQL - кеширование в Redis - редирект + отправка события.
3. Фоновый consumer читает события из Kafka и сохраняет в таблицу `click_events` (включая страну по IP).

## Запуск проекта
### Получение базы GeoIP (опционально)
Для определения страны по IP необходимо скачать бесплатную базу `GeoLite2-Country.mmdb` с сайта [MaxMind](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) (требуется регистрация). Поместите файл в папку `data/` в корне проекта. Если файл отсутствует, геолокация работать не будет, но остальные функции останутся доступны.

### Требования
- Docker и Docker Compose установлены.

### Шаги
1. Клонировать репозиторий:
   ```bash
   git clone https://github.com/yourusername/url-shortener-service.git
   cd url-shortener-service

(Опционально) Поместить файл базы GeoIP GeoLite2-Country.mmdb в папку data/. Если файла нет, геолокация работать не будет, но сервис останется работоспособным.

Запустить все сервисы:
docker-compose up -d
Приложение будет доступно на http://localhost:8080.


Остановка всех сервисов:
docker-compose down

API Endpoints
POST /shorten
Создаёт короткую ссылку. Требуется API-ключ в заголовке X-API-Key (тестовый ключ: test_api_key_123).


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

GET /metrics
Эндпоинт для Prometheus (метрики в формате, понятном Prometheus).

Примеры использования:
# Создать короткую ссылку
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test_api_key_123" \
  -d '{"url":"https://google.com"}'

# Перейти по короткой ссылке (в браузере)
http://localhost:8080/r/abc123

Переменные окружения
Все настройки задаются через переменные окружения (см. docker-compose.yml):

Переменная	Описание	По умолчанию
DB_HOST	Хост PostgreSQL	postgres
DB_PORT	Порт PostgreSQL	5432
DB_USER	Пользователь БД	admin
DB_PASSWORD	Пароль БД	securepassword123
DB_NAME	Имя базы данных	shortener_db
REDIS_HOST	Хост Redis	redis
REDIS_PORT	Порт Redis	6379
REDIS_PASSWORD	Пароль Redis	redispassword123
REDIS_DB	Номер базы Redis	0
KAFKA_BROKER	Адрес брокера Kafka	kafka:9092
KAFKA_TOPIC	Топик для событий	clicks
SERVER_PORT	Порт HTTP-сервера	8080


Планы по улучшению
Добавить веб-интерфейс.
Реализовать удаление и редактирование ссылок.
Добавить кэширование статистики.
Улучшить обработку ошибок и логирование.
CI/CD с GitHub Actions (автоматическая сборка и публикация образа).
CI/CD с GitHub Actions.


Лицензия
MIT (или другая)
