package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	brokerHost = "test.mosquitto.org"
	brokerPort = 1884
	username   = "rw"
	password   = "readwrite"
	clientId   = "texol-ble-driver"

	subTopic = "texol/ble/rawdata"
	pubTopic = "texol/ble/jsondata"
)

func onMessageReceived(client mqtt.Client, message mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())

	// customize code
	token := client.Publish(pubTopic, 0, false, message.Payload())
	token.Wait()

}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", brokerHost, brokerPort))
	opts.SetClientID(clientId)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetAutoReconnect(true)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := client.Subscribe(subTopic, 0, onMessageReceived); token.Wait() && token.Error() != nil {
		panic(fmt.Sprintf("Error subscribing to topic:", token.Error()))
	}

	fmt.Println("Subscribed to topic:", subTopic)

	// Wait for a signal to exit the program gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Drvier shutdown ...")
	client.Unsubscribe(subTopic)
	client.Disconnect(250)
}
