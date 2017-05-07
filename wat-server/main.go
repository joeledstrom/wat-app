package main

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	"log"

	pb "github.com/joeledstrom/wat-app/wat-api"
	"github.com/golang/protobuf/ptypes/timestamp"
)


type watServer struct {}


func (*watServer) ChatConnection(stream pb.Wat_ChatConnectionServer) error {

	for {
		msg, err := stream.Recv()
		if err != nil {
			break
		}

		stream.Send(&pb.ServerMessage{msg, &timestamp.Timestamp{}})

		log.Printf("Returned msg to client: %+v \n", msg)
	}

	return nil
}


func main() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9595))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()

	pb.RegisterWatServer(server, new(watServer))

	server.Serve(lis)

}
