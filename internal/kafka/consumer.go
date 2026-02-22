package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/services"
	"portfolio-rebalancer/internal/storage"
	"portfolio-rebalancer/internal/utils"
	"time"

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

		allocHash := utils.CanonicalHash(portfolio.NewAllocation)

		//get existing request or create new one
		p, err := storage.GetRebalanceRequest(ctx, portfolio.UserID)
		if err != nil {
			// Insert new request if not found
			if errors.Is(err, storage.ErrRequestNotFound) {
				rr := &models.RebalanceRequest{
					UserID:         portfolio.UserID,
					AllocationHash: allocHash,
				}
				if err = storage.SaveRebalanceRequest(ctx, rr); err != nil {
					log.Printf("Failed to save rebalance request: %v", err)
					return
				}

				p = rr
			} else {
				log.Printf("Failed to get current rebalance request: %v", err)
				return
			}
		} else {
			// due to open ended implementation of RebalanceTransaction,
			// RebalanceRequest only check idempotency based on allocation hash
			if p.AllocationHash == allocHash {
				log.Printf("No allocation changes detected for user: %s\n", portfolio.UserID)
				return
			}
		}

		log.Printf("Processing rebalance for user: %s\n", portfolio.UserID)

		transactions := services.CalculateRebalance(
			portfolio.UserID,
			portfolio.NewAllocation,
			portfolio.CurrentAllocation,
		)
		if len(transactions) > 0 {
			log.Printf("Saving %d transactions for user: %s\n", len(transactions), portfolio.UserID)
			// Retry mechanism with exponential backoff
			maxRetries := 5
			backoff := 1 * time.Second
			for i := 0; i < maxRetries; i++ {
				if err := storage.SaveRebalanceTransactions(ctx, transactions); err != nil {
					log.Printf("Failed to save rebalance transactions (attempt %d/%d): %v\n", i+1, maxRetries, err)
					if i == maxRetries-1 {
						log.Printf("CRITICAL: Failed to save transactions for user %s after %d attempts. Data may be lost.\n", portfolio.UserID, maxRetries)
					} else {
						time.Sleep(backoff)
						backoff *= 2
					}
				} else {
					break
				}
			}
		} else {
			log.Printf("No transactions to save for user: %s\n", portfolio.UserID)
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
