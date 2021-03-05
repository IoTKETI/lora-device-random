package main

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"math/rand"
	"fmt"
	"time"
	"strconv"
)

func main() {
	opts := mqtt.NewClientOptions().AddBroker("localhost:1883")
	var MsgHandler mqtt.MessageHandler = func(client mqtt.Client, mqtt_msg mqtt.Message) {
		fmt.Println("TOPIC: "+mqtt_msg.Topic())
		fmt.Printf("MSG : %s\n",mqtt_msg.Payload())
	}

	opts.SetDefaultPublishHandler(MsgHandler)
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := mqttClient.Subscribe("edgex",0,nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return
	}

	for {
		mqttClient.Publish("device",0,false, strconv.Itoa(rand.Intn(100)))
		time.Sleep(time.Second * 5)
	}
}
