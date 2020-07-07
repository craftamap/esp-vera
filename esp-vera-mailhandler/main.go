package main

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jordan-wright/email"
	. "github.com/logrusorgru/aurora"
)

var (
	mqttHost     = flag.String("mqtt-host", "tcp://192.168.0.0:1883", "")
	mqttUser     = flag.String("mqtt-user", "USERNAME", "")
	mqttPassword = flag.String("mqtt-password", "PASSWORD", "")

	mailFrom           = flag.String("mail-from", "espvera@mailhog.com", "")
	mailHost           = flag.String("mail-server-host", "localhost", "")
	mailPort           = flag.Int("mail-server-port", 1025, "")
	mailServerUser     = flag.String("mail-server-user", "espvera@mailhog.com", "")
	mailServerPassword = flag.String("mail-server-password", "", "")
	mailTo             = flag.String("mail-to", "your@email-address.com", "")
)

func sendMail() error {
	log.Println(Cyan(" :: "), "Sending eMail...")

	e := email.NewEmail()
	e.From = fmt.Sprintf("%s", *mailFrom)
	e.To = []string{
		*mailTo,
	}
	e.Subject = "Water me!"
	e.Text = []byte("ESP Vera needs some water")
	err := e.Send(fmt.Sprintf("%s:%d", *mailHost, *mailPort),
		smtp.PlainAuth("", *mailServerUser, *mailServerPassword, *mailHost))
	if err != nil {
		log.Println(Red(" :: "), "Mail could not be sent: ", err)
	}
	return err
}

func getMqttClient() mqtt.Client {

	opts := mqtt.NewClientOptions()
	opts.AddBroker(*mqttHost)
	opts.SetUsername(*mqttUser)
	opts.SetPassword(*mqttPassword)
	client := mqtt.NewClient(opts)
	return client
}

func handleMqttChannel(client mqtt.Client) chan string {
	c := make(chan string)
	go func() {
		client.Subscribe("/plants/moisture/1", 0, func(client mqtt.Client, msg mqtt.Message) {
			payload := string(msg.Payload())
			log.Println(Blue(" :: "), "Got payload:", payload)
			c <- string(payload)
		})
	}()
	return c

}

func main() {
	flag.Parse()
	log.Println(Green(" :: "), "Starting Esp-Vera Mailhandler")
	log.Println(Green(" :: "), "Connecting to MQTT Broker...")
	client := getMqttClient()
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(Red(" :: "), err)
	}
	log.Println(Green(" :: "), "Connected!")

	mailSent := false
	mailUpperThreshold := 90.0
	mailLowerThreshold := 82.0

	for payload := range handleMqttChannel(client) {
		f, err := strconv.ParseFloat(payload, 64)
		if err != nil {
			log.Print(Yellow(" :: "), err)
			continue
		}
		// Only send the mail if it was not already send
		// Only start resending if the value was higher than the upperThreshold once
		if f < mailLowerThreshold && !mailSent {
			err := sendMail()
			if err == nil {
				mailSent = true
				log.Println(Green(" :: "), "mailSent is true now")
			}
		} else if f > mailUpperThreshold && mailSent {
			mailSent = false
			log.Println(Green(" :: "), "mailSent is false now")
		}

	}
}
