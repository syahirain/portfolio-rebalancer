package kafka

import (
	"context"
	"encoding/json"
	"log"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/services"
	"portfolio-rebalancer/internal/storage"

	"github.com/segmentio/kafka-go"
)

func StartRebalanceConsumer(ctx context.Context) {
	err := ConsumeMessage(ctx, func(msg kafka.Message) {
		log.Printf("Received message: %s\n", string(msg.Value))

		if !isValidJSON(msg.Value) {
			log.Printf("Invalid JSON message, skipping: %s\n", string(msg.Value))
			return
		}

		var portfolio models.RebalancePortfolioKafka
		if err := json.Unmarshal(msg.Value, &portfolio); err != nil {
			log.Printf("Failed to unmarshal message: %v\n", err)
			return
		}

		log.Printf("Processing rebalance for user: %s\n", portfolio.UserID)

		transactions := services.CalculateRebalance(
			portfolio.UserID,
			portfolio.NewAllocation,
			portfolio.CurrentAllocation,
		)
		for _, tx := range transactions {
			log.Printf("Transaction: %v\n", tx)

			if err := storage.SaveRebalanceTransaction(ctx, tx); err != nil {
				log.Printf("Failed to save rebalance transaction: %v\n", err)
			}
		}
	})
	if err != nil {
		log.Printf("Failed to start consumer: %v\n", err)
	}
}

func isValidJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}
