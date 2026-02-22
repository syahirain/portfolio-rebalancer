package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

// InitKafka initializes kafka connection
func InitKafka() error {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	topic := os.Getenv("KAFKA_TOPIC")

	if kafkaBroker == "" || topic == "" {
		return nil // skip if env not set
	}

	if err := ensureTopicExists(kafkaBroker, topic, 1, 1); err != nil {
		return fmt.Errorf("failed to ensure topic exists: %w", err)
	}

	writer = &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	// Retry logic to check Kafka availability
	for i := 0; i < 10; i++ {
		err := writer.WriteMessages(context.Background(), kafka.Message{
			Value: []byte("ping"),
		})
		if err == nil {
			log.Println("Kafka is ready")
			return nil
		}
		log.Println("Waiting for Kafka to be ready...")
		time.Sleep(2 * time.Second)
	}

	return nil
}

func PublishMessage(ctx context.Context, payload []byte) error {
	if writer == nil {
		log.Println("Kafka writer is nil; skipping message publish")
		return fmt.Errorf("kafka writer not initialized")
	}

	msg := kafka.Message{
		Value: payload,
	}

	return writer.WriteMessages(ctx, msg)
}

func ConsumeMessage(ctx context.Context, handler func(kafka.Message)) error {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	topic := os.Getenv("KAFKA_TOPIC")

	if kafkaBroker == "" || topic == "" {
		log.Println("Kafka consumer config not set; skipping consumer start.")
		return nil
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaBroker},
		Topic:     topic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	reader.SetOffset(kafka.FirstOffset)

	go func() {
		defer reader.Close()
		for {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("Context canceled, stopping consumer")
					return
				}
				log.Printf("Kafka read error: %v\n", err)
				continue
			}

			handler(msg)
		}
	}()

	log.Println("Kafka consumer started")
	return nil
}

func ensureTopicExists(broker, topic string, partitions, replication int) error {
	// Connect to any broker first
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka broker: %w", err)
	}
	defer conn.Close()

	// Get controller (responsible for creating topics)
	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	// Check if topic already exists
	topics, err := controllerConn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("failed to read partitions: %w", err)
	}
	for _, t := range topics {
		if t.Topic == topic {
			log.Printf("Topic %s already exists", topic)
			return nil
		}
	}

	// Create topic if missing
	topicConfig := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replication,
		},
	}
	if err := controllerConn.CreateTopics(topicConfig...); err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	log.Printf("Topic %s created successfully", topic)
	return nil
}
