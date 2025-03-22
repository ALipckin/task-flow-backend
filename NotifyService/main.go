package main

import (
	"NotifyService/consumers"
	"NotifyService/initializers"
)

func init() {
	initializers.InitKafka()
}

func main() {

	defer initializers.Reader.Close()

	consumers.StartKafkaConsumer()
}
