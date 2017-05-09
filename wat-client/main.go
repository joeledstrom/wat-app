package main

import (
	"flag"
	"fmt"

	wat "github.com/joeledstrom/wat-app/wat-client-api-lib"
)


var (
	nick = flag.String("nick", "", "Nickname")
)

func main() {
	flag.Parse()

	if (*nick == "") {
		flag.PrintDefaults()
		return
	}

	client := wat.NewClient()

	err := client.Connect("127.0.0.1:9595", *nick)

	if (err != nil) {
		if _, ok := err.(*wat.NickAlreadyInUse); ok {
			fmt.Println("Nick Already in use. Try another nick")
		} else {
			fmt.Printf("Error %s", err)
		}
		return

	}

	sendChannel := make(chan string)
	recvChannel := make(chan string)

	go func() {
		for {
			content := <-sendChannel
			msg := wat.ClientMessage{Content: content}
			err := client.SendMessage(msg)
			if err != nil {
				break
			}
		}
	}()

	go func() {
		for {
			msg, err := client.RecvMessage()
			if err != nil {
				break
			}
			recvChannel <- (msg.Nick + ": " + msg.Content)
		}
	}()


	RunUi(sendChannel, recvChannel)
}

