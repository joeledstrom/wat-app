package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/google/uuid"
	pb "github.com/joeledstrom/wat-app/wat-api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type watClient struct {
	sessionToken string
	nick         string
	city         string
	loc          string
	outChannel   chan<- *pb.ServerMessage
}

type watServer struct {
	clientsMutex   sync.Mutex
	clientsByToken map[string]*watClient
	clientsByNick  map[string]*watClient
}

func newWatServer() *watServer {
	s := &watServer{}

	s.clientsMutex = sync.Mutex{}
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	s.clientsByToken = make(map[string]*watClient)
	s.clientsByNick = make(map[string]*watClient)

	return s
}

func (s *watServer) broadcastMessage(msg *pb.ServerMessage) {

	log.Printf("Broadcasting message: %+v \n", msg)

	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	for _, c := range s.clientsByToken {
		if c.outChannel != nil {
			c.outChannel <- msg
		}
	}
}

func (s *watServer) RegisterClient(_ context.Context, reg *pb.Registration) (*pb.RegistrationResponse, error) {

	_, found := s.clientsByNick[reg.Nick]

	if found {
		return &pb.RegistrationResponse{Status: pb.RegistrationResponse_NICK_ALREADY_IN_USE}, nil
	} else {
		client := &watClient{
			sessionToken: uuid.New().String(),
			nick:         reg.Nick,
			city:         reg.Location.City,
			loc:          reg.Location.Loc,
		}

		s.clientsMutex.Lock()
		defer s.clientsMutex.Unlock()
		s.clientsByNick[client.nick] = client
		s.clientsByToken[client.sessionToken] = client

		return &pb.RegistrationResponse{Status: pb.RegistrationResponse_OK, Token: client.sessionToken}, nil
	}

}

func (s *watServer) setOutChannelAndReturnClient(token string, outChannel chan *pb.ServerMessage) (watClient, bool) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	client, found := s.clientsByToken[token]

	if !found {
		return watClient{}, false
	}

	client.outChannel = outChannel

	return *client, true
}

func (s *watServer) OpenChat(stream pb.Wat_OpenChatServer) error {

	md, ok := metadata.FromIncomingContext(stream.Context())

	if ok {
		tokenMd, ok := md["session-token"]

		if ok && len(tokenMd) == 1 {

			token := tokenMd[0]
			outChannel := make(chan *pb.ServerMessage)
			client, ok := s.setOutChannelAndReturnClient(token, outChannel)

			if ok {

				// start sending goroutine
				go func() {
					for msg := range outChannel {
						err := stream.Send(msg)

						// if a message cant be sent to the client
						if err != nil {
							log.Printf("Removing client: %s \n", client.nick)

							// remove client
							s.clientsMutex.Lock()
							defer s.clientsMutex.Unlock()
							delete(s.clientsByToken, token)
							delete(s.clientsByNick, client.nick)

							break
						}
					}
				}()

				// start receive loop
				for {

					msg, err := stream.Recv()
					if err != nil {
						log.Printf("Lost connection to: %s \n", client.nick)
						break
					}
					serverMsg := &pb.ServerMessage{
						Nick:    client.nick,
						Content: msg.Content,
					}

					// only attach location when a client sends a "bot command"
					if strings.HasPrefix(msg.Content, "!") {
						serverMsg.Location = &pb.Location{client.city, client.loc}
					}

					s.broadcastMessage(serverMsg)
				}

				return nil
			}
		}
	}

	return errors.New("valid sessionToken missing")
}

var (
	port = flag.Int("port", 9595, "Listen on port")
)

func main() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	server := grpc.NewServer()

	pb.RegisterWatServer(server, newWatServer())

	log.Printf("wat-server started on port: %d\n", *port)

	server.Serve(lis)
}
