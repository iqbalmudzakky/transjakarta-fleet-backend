package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	amqp "github.com/rabbitmq/amqp091-go"

	"transjakarta-fleet-backend/internal/db"
	"transjakarta-fleet-backend/internal/geofence"
	"transjakarta-fleet-backend/internal/models"
	"transjakarta-fleet-backend/internal/repository"
)

type GeoFenceEvent struct {
	VehicleID string `json:"vehicle_id"`
	Event     string `json:"event"`
	Location  struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	Timestamp int64 `json:"timestamp"`
}

func main() {
	dbConn, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer dbConn.Close()

	repo := repository.NewLocationRepository(dbConn)

	mqttURL := os.Getenv("MQTT_BROKER_URL")
	rabbitURL := os.Getenv("RABBITMQ_URL")

	geoLat, _ := strconv.ParseFloat(os.Getenv("GEOFENCE_LAT"), 64)
	geoLng, _ := strconv.ParseFloat(os.Getenv("GEOFENCE_LNG"), 64)
	geoRadius, _ := strconv.ParseFloat(os.Getenv("GEOFENCE_RADIUS"), 64)

	// RabbitMQ Connection
	rabbitConn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare exchange
	if err := ch.ExchangeDeclare("fleet.events", "fanout", true, false, false, false, nil); err != nil {
		log.Fatalf("failed to declare exchange: %v", err)
	}

	// Declare queue
	q, err := ch.QueueDeclare("geofence_alerts", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	if err := ch.QueueBind(q.Name, "", "fleet.events", false, nil); err != nil {
		log.Fatalf("failed to bind queue: %v", err)
	}

	// MQTT setup
	opts := mqtt.NewClientOptions().AddBroker(mqttURL)
	opts.SetClientID("fleet-subscriber")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to connect to MQTT broker: %v", token.Error())
	}

	topic := "/fleet/vehicle/+/location"

	if token := client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		var loc models.VehicleLocation
		if err := json.Unmarshal(msg.Payload(), &loc); err != nil {
			log.Printf("failed to unmarshal payload: %v", err)
			return
		}

		if err := repo.InsertLocation(&loc); err != nil {
			log.Printf("failed to insert location: %v", err)
			return
		}

		// Check geofence
		d := geofence.DistanceInMeters(loc.Latitude, loc.Longitude, geoLat, geoLng)
		log.Printf("Distance from geofence: %.2f meters", d)
		if d <= geoRadius {
			var event GeoFenceEvent
			event.VehicleID = loc.VehicleID
			event.Event = "geofence_entry"
			event.Location.Latitude = loc.Latitude
			event.Location.Longitude = loc.Longitude
			event.Timestamp = loc.Timestamp

			body, err := json.Marshal(event)
			if err != nil {
				log.Printf("failed to marshal event: %v", err)
				return
			}

			if err := ch.Publish("fleet.events", "", false, false, amqp.Publishing{
				ContentType: "application/json",
				Body:        body,	
			}); err != nil {
				log.Printf("failed to publish event: %v", err)
			} else {
				log.Printf("geofence event sent for %s", loc.VehicleID)
			}
		}
	}); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to subscribe: %v", token.Error())
	}

	log.Println("MQTT subscriber running...")
	select {} // block forever
}