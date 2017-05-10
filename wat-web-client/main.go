package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"net"
	"fmt"
	"log"
	"flag"
	"time"
	wat "github.com/joeledstrom/wat-app/wat-client-api-lib"
)

var (
	port = flag.Int("port", 8080, "Serve web client on a HTTP server at this port")
	watServerHost = flag.String("wat-server-host", "", "Hostname or IP of wat-server")
	watServerPort = flag.Int("wat-server-port", 9595, "Port of wat-server")
)

func main() {

	http.Handle("/", http.FileServer(http.Dir("./web-client-angular/dist")))
	http.HandleFunc("/ws", onWsConnection)

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	log.Printf("wat-web-client http server started on port: %d\n", *port)

	http.Serve(listen, nil)
}


var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

type message struct {
	Nick string
	Content string
}

func onWsConnection(w http.ResponseWriter, r *http.Request) {
	log.Println("opening websocket connection...")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to a websocket connection:", err)
		return
	}
	defer ws.Close()


	// first message is a dummy msg containing only the requested nickname
	msg := &message{}
	err = ws.ReadJSON(msg)
	if err != nil {
		log.Println("Error reading 'dummy' message:", err)
		return
	}
	nick := msg.Nick


	client := wat.NewClient()
	defer client.Close()
	err = client.Connect(fmt.Sprintf("%s:%d", *watServerHost, *watServerPort), nick)

	if err != nil {
		log.Println("Failed to connect to the wat-server:", err)
		return
	}

	go func() {
		for {
			watMsg, err := client.RecvMessage()

			if err != nil {
				log.Println("Error receiving from wat-server:", err)
				break;
			}

			msg := &message{Content:watMsg.Content, Nick:watMsg.Nick}
			err = ws.WriteJSON(msg)

			if err != nil {
				log.Println("Error sending message:", err)
				break;
			}
		}

	}()

	for {
		err := ws.ReadJSON(msg)

		if err != nil {
			log.Println("Error reading message:", err)
			break;
		}

		err = client.SendMessage(wat.ClientMessage{Content: msg.Content})

		if err != nil {
			log.Println("Error sending to wat-server:", err)
			break;
		}
	}
}