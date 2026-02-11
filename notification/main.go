package main

import (
	"context"
	"log"
	"notification/consumers"
	"notification/initializers"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	initializers.InitKafka()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		if initializers.Reader != nil {
			log.Println("Closing Kafka reader...")
			if err := initializers.Reader.Close(); err != nil {
				log.Printf("Error closing Kafka reader: %v", err)
			}
		}
		if initializers.Writer != nil {
			log.Println("Closing Kafka writer...")
			if err := initializers.Writer.Close(); err != nil {
				log.Printf("Error closing Kafka writer: %v", err)
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		consumers.StartKafkaConsumer(ctx)
	}()

	sig := <-sigChan
	log.Printf("Received signal: %v, initiating graceful shutdown...", sig)
	cancel()

	log.Println("notification shutdown complete")
}
