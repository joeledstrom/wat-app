package main

import (
	"fmt"
	"sync"
	"net"
	"log"

	pb "github.com/joeledstrom/wat-app/wat-api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)



type client struct {
	token uuid.UUID
	nick string
	ip string
	outChannel chan<- *pb.ServerMessage
}



type watServer struct {
	clientsMutex	sync.Mutex
	clientsByToken 	map[uuid.UUID]*client
	clientsByNick	map[string]*client
}

func newWatServer() *watServer {
	s := &watServer{}

	s.clientsMutex = sync.Mutex{}
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	s.clientsByToken = make(map[uuid.UUID]*client)
	s.clientsByNick = make(map[string]*client)


	return s
}

func (s *watServer) broadcastMessage(msg *pb.ServerMessage) {

	log.Printf("Broadcasting message: %+v \n", msg)

	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	for _, c := range s.clientsByToken {
		if (c.outChannel != nil) {
			c.outChannel <- msg
		}
	}



}


func (s *watServer) RegisterClient(_ context.Context, reg *pb.Registration) (*pb.RegistrationResponse, error) {

	_, found := s.clientsByNick[reg.Nick]

	if found {
		return &pb.RegistrationResponse{Status: pb.RegistrationResponse_NICK_ALREADY_IN_USE}, nil

	} else {
		c := &client{}


		c.token = uuid.New()
		c.nick = reg.Nick
		c.ip = reg.Ip

		s.clientsMutex.Lock()
		defer s.clientsMutex.Unlock()
		s.clientsByNick[c.nick] = c
		s.clientsByToken[c.token] = c


		return &pb.RegistrationResponse{Status: pb.RegistrationResponse_OK, Token: c.token.String()}, nil
	}

}



func (s *watServer) setOutChannel(token uuid.UUID, outChannel chan *pb.ServerMessage) (string, string, bool) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	c, found := s.clientsByToken[token]

	if !found {
		return "", "", false
	}

	c.outChannel = outChannel

	return c.nick, c.ip, true
}


func (s *watServer) OpenChat(stream pb.Wat_OpenChatServer) error {

	md, ok := metadata.FromIncomingContext(stream.Context())

	if ok  {
		tokens, ok := md["token"]

		if ok && len(tokens) == 1 {

			token := tokens[0]
			t, _ := uuid.Parse(token)
			outChannel := make(chan *pb.ServerMessage)

			nick, ip, ok := s.setOutChannel(t, outChannel)


			if ok {

				// start sending goroutine
				go func() {
					for msg := range outChannel {
						err := stream.Send(msg)

						// if a message cant be sent to the client
						if err != nil {
							log.Printf("Removing client: %s \n", nick)

							// remove client
							s.clientsMutex.Lock()
							defer s.clientsMutex.Unlock()
							delete(s.clientsByToken, t)
							delete(s.clientsByNick, nick)

							break;
						}
					}
				}()

				// start receive loop
				for {

					msg, err := stream.Recv()
					if err != nil {
						log.Printf("Lost connection to: %s \n", nick)
						break
					}
					serverMsg := &pb.ServerMessage{
						Nick: nick,
						Content: msg.Content,
						Ip: ip,
					}

					s.broadcastMessage(serverMsg)
				}

				return nil
			}



		}

	}

	return nil  // should return an error here
}


func main() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9595))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()

	pb.RegisterWatServer(server, newWatServer())

	server.Serve(lis)

}
