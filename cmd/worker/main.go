package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	_ = godotenv.Load()
	rabbitURL := os.Getenv("RABBITMQ_URL")

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("geofence_alerts", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to consume queue: %v", err)
	}

	log.Println("Worker listening for geofence events...")
	
	for msg := range msgs {
		log.Printf("Received geofence event: %s", msg.Body)
	}
}