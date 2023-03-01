package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type MQTTMessage struct {
	Topic   string `json:"topic"`
	Payload []byte `json:"payload"`
	QoS     uint8  `json:"qos"`
	Retain  bool   `json:"retain"`
}

var DbDsn = ""
var MqttServer = ""

var MqttUsername = ""
var MqttPassword = ""

func main() {

	var rootCmd = &cobra.Command{
		Use:   "mqtt2mysql",
		Short: "MQTT to MySQL data pipeline",
		Long: `Data processing solution that subscribes to all topics on an MQTT broker, receives messages published on those topics, and stores them in a MySQL database
Complete documentation is available at https://github.com/yuksbg/mqtt2mysql`,
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
	rootCmd.PersistentFlags().StringVarP(&DbDsn, "db", "d", "user:pass@tcp(127.0.0.1:3306/app)", "MySQL DSN")
	rootCmd.PersistentFlags().StringVarP(&MqttServer, "mqtt_server", "m", "tcp://127.0.0.1:1883", "MQTT DSN")
	rootCmd.PersistentFlags().StringVarP(&MqttUsername, "mqtt_username", "u", "", "MQTT Username")
	rootCmd.PersistentFlags().StringVarP(&MqttPassword, "mqtt_password", "p", "", "MQTT Password")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() {
	db, err := sqlx.Connect("mysql", DbDsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	mqttOpts := mqtt.NewClientOptions().AddBroker(MqttServer)
	mqttOpts.SetClientID("mqtt-to-db")
	if MqttUsername != "" {
		mqttOpts.SetUsername(MqttUsername)
	}
	if MqttPassword != "" {
		mqttOpts.SetPassword(MqttPassword)
	}

	client := mqtt.NewClient(mqttOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	} else {
		log.Println("Connected to MQTT Broker")
	}
	defer client.Disconnect(250)

	msgCh := make(chan MQTTMessage)
	client.Subscribe("#", 0, func(_ mqtt.Client, msg mqtt.Message) {
		msgCh <- MQTTMessage{
			Topic:   msg.Topic(),
			Payload: msg.Payload(),
			QoS:     msg.Qos(),
			Retain:  msg.Retained(),
		}
	})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case msg := <-msgCh:
			if _, err := db.NamedExec(`
				INSERT INTO mqtt (topic, value, qos, retain)
				VALUES (:topic, :payload, :qos, :retain)
				ON DUPLICATE KEY UPDATE value = :payload, qos = :qos, retain = :retain
			`, &msg); err != nil {
				log.Printf("Failed to insert mqtt message into mqtt table: %v", err)
			}

		case sig := <-sigCh:
			log.Printf("Received signal %v, shutting down...", sig)
			client.Disconnect(250)
			os.Exit(0)
		}
	}
}
