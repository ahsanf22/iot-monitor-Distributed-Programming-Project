package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var shouldRun = true
var temp = 21.0
var humidity = 45.0

var cmdHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	command := string(msg.Payload())
	if command == "STOP" {
		log.Println("COMMAND: Stopping Sensor...")
		shouldRun = false
	} else if command == "START" {
		log.Println("COMMAND: Starting Sensor...")
		shouldRun = true
	}
}

func main() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetClientID("go_sensor_1")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Mosquitto is not running.")
	}
	client.Subscribe("home/sensor1/cmd", 0, cmdHandler)
	log.Println("Sensor 1 Active (Temp + Humidity)...")

	for {
		if shouldRun {
			// Drifting Physics
			temp += (rand.Float64() - 0.5) * 0.4
			humidity += (rand.Float64() - 0.5) * 1.0
			
			tPayload := fmt.Sprintf("%.2f", temp)
			hPayload := fmt.Sprintf("%.1f", humidity)
			
			client.Publish("home/sensor1/temp", 0, false, tPayload)
			client.Publish("home/sensor1/humid", 0, false, hPayload)
			
			log.Printf("Sensor 1 -> %s°C | %s%%", tPayload, hPayload)
		}
		time.Sleep(2 * time.Second)
	}
}
