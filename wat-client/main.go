package main

import (
	"flag"
	"log"
	"context"

	pb "github.com/joeledstrom/wat-app/wat-api"
	"google.golang.org/grpc"
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


	conn, err := grpc.Dial("127.0.0.1:9595", grpc.WithInsecure())
	if err != nil {
		log.Panicln(err)
	}

	client := pb.NewWatClient(conn)

	chatConn, err := client.ChatConnection(context.Background())
	if err != nil {
		log.Panicln(err)
	}

	sendChannel := make(chan string)
	recvChannel := make(chan string)

	go func() {
		for {
			content := <-sendChannel
			msg := pb.ClientMessage{Sender: *nick, Content: content}
			err = chatConn.Send(&msg)
			if err != nil {
				break
			}
		}
	}()

	go func() {
		for {
			msg, err := chatConn.Recv()
			if err != nil {
				break
			}
			recvChannel <- msg.Message.Content
		}
	}()


	RunUi(sendChannel, recvChannel)
}

