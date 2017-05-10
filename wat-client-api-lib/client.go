package wat_client_api_lib

import (
	"errors"
	"net/http"
	"time"
	"encoding/json"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	pb "github.com/joeledstrom/wat-app/wat-api"
)

type Client interface {
	Connect(address, nick string) error
	SendMessage(msg ClientMessage) error
	RecvMessage() (*ServerMessage, error)
	Close()
}

type ClientMessage struct {
	Content string
}

type ServerMessage struct {
	Nick string
	Content string
	Location *Location
}

type Location struct {
	City string
	Loc string
}


type NickAlreadyInUse struct {}

func (*NickAlreadyInUse) Error() string {
	return "Nick already in use"
}

func (c *client) Connect(address, nick string) (e error) {

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	c.conn = conn

	defer func() {
		if e != nil {
			conn.Close()
		}
	}()


	if err != nil {
		return err
	}

	watClient := pb.NewWatClient(conn)


	loc, err := getLocation()
	if err != nil {
		return err
	}

	reg := &pb.Registration{Nick:nick, Location:loc}

	resp, err := watClient.RegisterClient(context.Background(), reg)

	if err != nil {
		return err
	}

	if resp.Status == pb.RegistrationResponse_NICK_ALREADY_IN_USE {
		return new(NickAlreadyInUse)
	}

	md := metadata.Pairs("session-token", resp.Token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	c.chatConn, err = watClient.OpenChat(ctx)

	return err
}

func (c *client) SendMessage(msg ClientMessage) error {
	if c.chatConn == nil {
		return errors.New("Not Connected. Call .Connect(addr, nick) first.")
	}

	pbMsg := pb.ClientMessage{Content: msg.Content}
	return c.chatConn.Send(&pbMsg)
}

func (c *client) RecvMessage() (*ServerMessage, error) {
	if c.chatConn == nil {
		return nil, errors.New("Not Connected. Call .Connect(addr, nick) first.")
	}

	pbMsg, err := c.chatConn.Recv()

	if err != nil {
		return nil, err
	}

	msg := &ServerMessage{
		Nick: pbMsg.Nick,
		Content: pbMsg.Content,
	}

	if pbMsg.Location != nil {
		msg.Location = &Location{pbMsg.Location.City, pbMsg.Location.Loc}
	}

	return msg, nil
}

func (c*client) Close() {
	c.conn.Close()
}


func NewClient() Client {
	return &client{}
}

type client struct {
	conn		*grpc.ClientConn
	chatConn 	pb.Wat_OpenChatClient
}

func getLocation() (*pb.Location, error) {
	httpClient := &http.Client{Timeout: 5 * time.Second}

	req, _ := http.NewRequest("GET", "http://ipinfo.io", nil)
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	var loc pb.Location
	err = json.NewDecoder(resp.Body).Decode(&loc)

	if err != nil {
		return nil, err
	}

	return &loc, nil
}




