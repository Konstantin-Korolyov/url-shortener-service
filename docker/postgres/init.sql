-- Создаём таблицу для пользователей (если будем делать авторизацию)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    api_key VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Основная таблица: оригинальные ссылки
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    clicks INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Индексы для быстрого поиска
CREATE INDEX idx_short_code ON urls(short_code);
CREATE INDEX idx_user_id ON urls(user_id);
CREATE INDEX idx_created_at ON urls(created_at);

-- Таблица для детальной статистики (будем заполнять через Kafka)
CREATE TABLE IF NOT EXISTS click_events (
    id SERIAL PRIMARY KEY,
    url_id INTEGER REFERENCES urls(id) ON DELETE CASCADE,
    ip_address INET,
    user_agent TEXT,
    referer TEXT,
    country_code CHAR(2),
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска статистики
CREATE INDEX idx_url_id_time ON click_events(url_id, clicked_at);
CREATE INDEX idx_clicked_at ON click_events(clicked_at);

-- Вставляем тестового пользователя
INSERT INTO users (email, api_key) 
VALUES ('test@example.com', 'test_api_key_123')
ON CONFLICT (email) DO NOTHING;

-- Вставляем тестовую ссылку
INSERT INTO urls (original_url, short_code, user_id) 
VALUES ('https://github.com', 'gh', 1)
ON CONFLICT (short_code) DO NOTHING;