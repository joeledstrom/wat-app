package main

import (
	"flag"
	"fmt"

	wat "github.com/joeledstrom/wat-app/wat-client-api-lib"
	"strings"
)



var (
	nick = flag.String("nick", "SMHI-WeatherBot", "Nickname")
	host = flag.String("host", "", "Hostname or IP")
	port = flag.Int("port", 9595, "Port")
)


func main() {
	flag.Parse()

	if (*host == "") {
		flag.PrintDefaults()
		return
	}

	client := wat.NewClient()

	err := client.Connect(fmt.Sprintf("%s:%d", *host, *port), *nick)

	if err != nil {
		if _, ok := err.(*wat.NickAlreadyInUse); ok {
			fmt.Println("Nick Already in use. Try another nick")
		} else {
			fmt.Printf("Error connecting: %s\n", err)
		}
		return

	}

	err = messageRecvLoop(client)

	fmt.Printf("Lost connection: %s\n", err)
}

func messageRecvLoop(client wat.Client) error {
	for {
		msg, err := client.RecvMessage()
		if err != nil {
			return err
		}

		if strings.HasPrefix(msg.Content, "!weather") {
			err := client.SendMessage(wat.ClientMessage{"Current weather..... TODO"})

			if err != nil {
				return err
			}
		}
	}
}