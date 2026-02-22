package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"portfolio-rebalancer/internal/kafka"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture Ctrl+C / SIGTERM
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Shutting down consumer...")
		cancel()
	}()

	// Optionally, ensure Kafka topic exists
	if err := kafka.InitKafka(); err != nil {
		log.Fatalf("Kafka init failed: %v", err)
	}

	// Start consuming messages
	kafka.StartRebalanceConsumer(ctx)

	// Keep running until context is canceled
	<-ctx.Done()
	log.Println("Consumer stopped")
}
