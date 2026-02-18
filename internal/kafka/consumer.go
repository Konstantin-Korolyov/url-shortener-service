package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

// ConsumerConfig содержит настройки для консюмера
type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// ClickEventConsumer читает события из Kafka и сохраняет в БД
type ClickEventConsumer struct {
	reader *kafka.Reader
	db     *pgxpool.Pool
}

// NewClickEventConsumer создаёт нового потребителя
func NewClickEventConsumer(cfg ConsumerConfig, db *pgxpool.Pool) *ClickEventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	return &ClickEventConsumer{reader: reader, db: db}
}

// Start запускает бесконечный цикл чтения сообщений
func (c *ClickEventConsumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.reader.Close()
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}
			// Обрабатываем сообщение
			if err := c.processMessage(ctx, msg); err != nil {
				log.Printf("Failed to process message: %v", err)
				// Здесь можно реализовать повторные попытки, но для простоты пропускаем
			} else {
				// Сообщение успешно обработано, коммитим
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					log.Printf("Failed to commit message: %v", err)
				}
			}
		}
	}
}

func (c *ClickEventConsumer) processMessage(ctx context.Context, msg kafka.Message) error {
	var event ClickEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	// Вставляем запись в таблицу click_events
	query := `
		INSERT INTO click_events (url_id, ip_address, user_agent, referer, country_code, clicked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	// country_code пока оставляем NULL, можно определить позже
	_, err := c.db.Exec(ctx, query,
		event.URLID,
		event.IP,
		event.UserAgent,
		event.Referer,
		nil, // country_code
		event.Timestamp,
	)
	return err
}
