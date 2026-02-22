package kafka

import (
	"context"
	"encoding/json"
	"log"

	// removed import to break cycle; internalKafka.ConsumeMessage is now called directly in this package
	"portfolio-rebalancer/internal/models"

	"github.com/segmentio/kafka-go"
)

func StartRebalanceConsumer(ctx context.Context) {
	err := ConsumeMessage(ctx, func(msg kafka.Message) {
		log.Printf("Received message: %s\n", string(msg.Value))

		if !isValidJSON(msg.Value) {
			log.Printf("Invalid JSON message, skipping: %s\n", string(msg.Value))
			return
		}

		var portfolio models.UpdatedPortfolio
		if err := json.Unmarshal(msg.Value, &portfolio); err != nil {
			log.Printf("Failed to unmarshal message: %v\n", err)
			return
		}

		log.Printf("Processing rebalance for user: %s\n", portfolio.UserID)
		// TODO: Fetch current allocation and call CalculateRebalance
	})
	if err != nil {
		log.Printf("Failed to start consumer: %v\n", err)
	}
}

func isValidJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}
