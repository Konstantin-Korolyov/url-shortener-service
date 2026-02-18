package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

// ClickEvent структура события клика (будет отправляться в Kafka)
type ClickEvent struct {
	URLID     int64     `json:"url_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referer   string    `json:"referer"`
	Timestamp time.Time `json:"timestamp"`
}

// Producer обёртка над kafka.Writer
type Producer struct {
	writer *kafka.Writer
}

// NewProducer создаёт нового продюсера для указанного топика
func NewProducer(brokers []string, topic string) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{}, // распределение сообщений по партициям
		RequiredAcks: kafka.RequireNone,   // не ждём подтверждения от всех реплик (быстрее)
		Async:        true,                // асинхронная отправка
	}
	return &Producer{writer: w}
}

// PublishClick отправляет событие клика в Kafka
func (p *Producer) PublishClick(ctx context.Context, event ClickEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Key:   []byte(event.IP), // опционально, можно использовать для партиционирования
		Value: data,
	}
	return p.writer.WriteMessages(ctx, msg)
}

// Close закрывает продюсер
func (p *Producer) Close() error {
	return p.writer.Close()
}
