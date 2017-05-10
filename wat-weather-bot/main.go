package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
	"log"

	wat "github.com/joeledstrom/wat-app/wat-client-api-lib"
)



var (
	nick = flag.String("nick", "SMHI-WeatherBot", "Nickname")
	host = flag.String("host", "", "Hostname or IP")
	port = flag.Int("port", 9595, "Port")
)


func main() {
	flag.Parse()

	if *host == "" {
		flag.PrintDefaults()
		return
	}


	for {
		err := messageRecvLoop()

		if err != nil {
			if _, ok := err.(*wat.NickAlreadyInUse); ok {
				fmt.Println("Nick Already in use. Try another nick")
				return
			}

		}

		log.Printf("Error: %s\n", err)
		log.Println("Retrying/reconnecting in 5")
		time.Sleep(5 * time.Second)
	}

}

func messageRecvLoop() error {
	client := wat.NewClient()
	defer client.Close()
	err := client.Connect(fmt.Sprintf("%s:%d", *host, *port), *nick)

	if err != nil {
		return err
	}

	for {
		log.Println("wat-weather-bot listening for messages")
		msg, err := client.RecvMessage()
		if err != nil {
			return err
		}

		if strings.HasPrefix(msg.Content, "!weather") {
			forecast := "Current temperature in " + msg.Location.City + ": " + "5 C"
			err := client.SendMessage(wat.ClientMessage{forecast})

			if err != nil {
				return err
			}
		}
	}
}