package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type VehicleLocation struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

func main() {
	mqttURL := os.Getenv("MQTT_BROKER_URL")
	if mqttURL == "" {
		mqttURL = "tcp://localhost:1883"
	}

	vehicleID := "B1234XYZ"

	opts := mqtt.NewClientOptions().AddBroker(mqttURL)
	opts.SetClientID("mock-publisher")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to connect to MQTT broker: %v", token.Error())
	}

	lat := -6.2088
	lng := 106.8456

	for {
		loc := VehicleLocation{
			VehicleID: vehicleID,
			Latitude:  lat + (rand.Float64()-0.5)*0.0002,
			Longitude: lng + (rand.Float64()-0.5)*0.0002,
			Timestamp: time.Now().Unix(),
		}

		payload, _ := json.Marshal(loc)
		topic := "/fleet/vehicle/" + vehicleID + "/location"

		if token := client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
			log.Printf("failed to publish message: %v", token.Error())
		} else {
			log.Printf("published location: %s", payload)
		}

		time.Sleep(2 * time.Second)
	}
}