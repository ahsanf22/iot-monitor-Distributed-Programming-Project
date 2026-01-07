package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var shouldRun = true
var temp = 24.0
var humidity = 50.0

var cmdHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	command := string(msg.Payload())
	if command == "STOP" {
		log.Println("🛑 COMMAND: Stopping Sensor...")
		shouldRun = false
	} else if command == "START" {
		log.Println("✅ COMMAND: Starting Sensor...")
		shouldRun = true
	}
}

func main() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetClientID("go_sensor_2")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("❌ Mosquitto is not running.")
	}
	client.Subscribe("home/sensor2/cmd", 0, cmdHandler)
	log.Println("✅ Sensor 2 Active (Temp + Humidity)...")

	for {
		if shouldRun {
			temp += (rand.Float64() - 0.5) * 0.4
			humidity += (rand.Float64() - 0.5) * 1.0
			
			tPayload := fmt.Sprintf("%.2f", temp)
			hPayload := fmt.Sprintf("%.1f", humidity)
			
			client.Publish("home/sensor2/temp", 0, false, tPayload)
			client.Publish("home/sensor2/humid", 0, false, hPayload)
			
			log.Printf("Sensor 2 -> %s°C | %s%%", tPayload, hPayload)
		}
		time.Sleep(2 * time.Second)
	}
}
