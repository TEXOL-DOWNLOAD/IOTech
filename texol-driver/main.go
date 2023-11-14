package main

import (
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/texol/texol-ble-driver/logfile"
)

var version = "1.1.0"
var GatewayIP string

const (
	Topic_bridgeJoin  = "/CLIENT"
	Topic_SensorData  = "/SENSOR/DATA"
	Topic_SensorAlive = "/SENSOR/ALIVE"
)

var EventChan = make(chan eventPro)
var mqttConfig = mqttCPro{
	ClientID:          "texol-ble-driver",
	IP:                "texol-broker", //192.168.0.21 //test.mosquitto.org //texol-broker"
	Port:              "1883",         //1883
	Username:          "",
	Password:          "",
	CAcertificateFile: "",
}

func init() {
	logfile.Println("Software Start")
	logfile.Println("Version: " + version)

	//Read Gateway IP File
	GatewayIP = readIPfile("/texol/GatewayIP.txt") ///texol/GatewayIP.txt
	logfile.Println("Gateway IP: " + GatewayIP)
}

/*
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
*/

func main() {
	/*
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
	*/
	defer logfile.Println("Software Close")

	//Connect MQTT Broker
	STopic := []string{Topic_bridgeJoin, Topic_SensorData, Topic_SensorAlive}
	go mQTTCon(mqttConfig, EventChan, STopic, 0)

	// Wait for a signal to exit the program gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	logfile.Println("Drvier shutdown ...")

	EventChan <- eventPro{
		Message: "Close",
		Value:   nil,
	}
	time.Sleep(1 * time.Second)
	/*
		client.Unsubscribe(subTopic)
		client.Disconnect(250)
	*/
}

func readIPfile(Path string) string {
	content, err := os.ReadFile(Path)
	if err != nil {
		logfile.Println(err)
		os.Exit(0)
	}

	IP := strings.Split(string(content), "\n")[0]
	if net.ParseIP(IP) == nil {
		logfile.Println("Gateway IP ERROR")
		os.Exit(0)
	}

	return IP
}
