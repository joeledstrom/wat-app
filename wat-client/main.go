package main

import (
	"flag"
	"log"
	"context"

	pb "github.com/joeledstrom/wat-app/wat-api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"fmt"
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
	defer conn.Close()

	client := pb.NewWatClient(conn)


	ip := "kaka"  // TODO: fetch this from http://bot.whatismyipaddress.com/

	reg := &pb.Registration{Nick:*nick, Ip:ip}

	resp, err := client.RegisterClient(context.Background(), reg)

	if err != nil {
		log.Panicln(err)
	}

	if (resp.Status == pb.RegistrationResponse_NICK_ALREADY_IN_USE) {
		fmt.Println("Nick Already in use. Try another nick")
		return
	}


	md := metadata.Pairs("session-token", resp.Token)

	ctx := metadata.NewOutgoingContext(context.Background(), md)


	chatConn, err := client.OpenChat(ctx)
	if err != nil {
		log.Panicln(err)
	}

	sendChannel := make(chan string)
	recvChannel := make(chan string)

	go func() {
		for {
			content := <-sendChannel
			msg := pb.ClientMessage{Content: content}
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
			recvChannel <- (msg.Nick + ": " + msg.Content)
		}
	}()


	RunUi(sendChannel, recvChannel)
}

