package main

import (
	"fmt"
	"sync"
	"net"
	"log"
	"errors"

	pb "github.com/joeledstrom/wat-app/wat-api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)


type client struct {
	sessionToken string
	nick         string
	ip           string
	outChannel   chan<- *pb.ServerMessage
}

type watServer struct {
	clientsMutex	sync.Mutex
	clientsByToken 	map[string]*client
	clientsByNick	map[string]*client
}

func newWatServer() *watServer {
	s := &watServer{}

	s.clientsMutex = sync.Mutex{}
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	s.clientsByToken = make(map[string]*client)
	s.clientsByNick = make(map[string]*client)

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
		c := &client{
			sessionToken: uuid.New().String(),
			nick: reg.Nick,
			ip: reg.Ip,
		}

		s.clientsMutex.Lock()
		defer s.clientsMutex.Unlock()
		s.clientsByNick[c.nick] = c
		s.clientsByToken[c.sessionToken] = c


		return &pb.RegistrationResponse{Status: pb.RegistrationResponse_OK, Token: c.sessionToken}, nil
	}

}



func (s *watServer) setOutChannelAndReturnClientInfo(token string, outChannel chan *pb.ServerMessage) (string, string, bool) {
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
		tokenMd, ok := md["session-token"]

		if ok && len(tokenMd) == 1 {

			token := tokenMd[0]
			outChannel := make(chan *pb.ServerMessage)
			nick, ip, ok := s.setOutChannelAndReturnClientInfo(token, outChannel)

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
							delete(s.clientsByToken, token)
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

	return errors.New("valid sessionToken missing")
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
