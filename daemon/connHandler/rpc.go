package connhandler

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	pb "github.com/Ahmed-Armaan/Localhost.git/proto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func requestListenerLoop(stream pb.TunnelService_HTTPTunnelClient, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("stream closed")
			} else {
				log.Printf("stream error: %v\n", err)
			}
			return
		}

		switch msg.GetType() {
		case pb.MessageType_DATA:
			fmt.Printf("Received request:%v\n", msg)
			sendResponse(stream)
		case pb.MessageType_HEARTBEAT:
			fmt.Println("Heartbeat received")
		}
	}
}

func newConnection(stream pb.TunnelService_HTTPTunnelClient) {
	msg := &pb.HTTPMessage{
		Type:   pb.MessageType_NEW_CONNECTION,
		ConnId: "client-123",
	}
	if err := stream.Send(msg); err != nil {
		log.Println("failed to send NEW_CONNECTION:", err)
	}
}

func sendResponse(stream pb.TunnelService_HTTPTunnelClient) {
	time.Sleep(2 * time.Second)

	res := &pb.HTTPResponseData{
		StatusCode: 200,
		StatusText: "OK",
		Headers: map[string]*pb.HeaderValues{
			"Content-Type": {Values: []string{"text/plain"}},
		},
		Body: []byte("Hello from remote app!"),
	}

	msg := &pb.HTTPMessage{
		Type:    pb.MessageType_DATA,
		ConnId:  "client-123",
		Payload: &pb.HTTPMessage_Response{Response: res},
	}

	if err := stream.Send(msg); err != nil {
		log.Println("failed to send response:", err)
	}
}

func grpcListener() {
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTunnelServiceClient(conn)
	stream, err := client.HTTPTunnel(context.Background())
	if err != nil {
		log.Fatalf("failed to start stream: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go requestListenerLoop(stream, &wg)
	newConnection(stream)

	wg.Wait()
}
