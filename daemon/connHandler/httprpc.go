package connhandler

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	pb "github.com/Ahmed-Armaan/Localhost.git/proto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func httpRequestListenerLoop(stream pb.TunnelService_HTTPTunnelClient, wg *sync.WaitGroup) {
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
		case pb.MessageType_NEW_CONNECTION:
			if msg.GetResponse().StatusCode != 200 {
				log.Fatal("Connection failed")
			} else {
				fmt.Printf("Connected: Access app at %s\n", msg.GetResponse().GetStatusText())
			}

		case pb.MessageType_DATA:
			fmt.Printf("Received request with connId%v\n:%v\n", msg.GetConnId(), msg)
			reqDataChan <- reqData{
				connId: msg.GetConnId(),
				appId:  msg.GetAppId(),
				req:    msg,
			}

		case pb.MessageType_HEARTBEAT:
			fmt.Println("Heartbeat received")
		}
	}
}

func newHttpConnection(stream pb.TunnelService_HTTPTunnelClient, appName string) {
	msg := &pb.HTTPMessage{
		Type:  pb.MessageType_NEW_CONNECTION,
		AppId: appName,
	}
	if err := stream.Send(msg); err != nil {
		log.Println("failed to send NEW_CONNECTION:", err)
	}
}

func sendHttpresponse(stream pb.TunnelService_HTTPTunnelClient, response *reqData) {
	if err := stream.Send(response.req); err != nil {
		log.Println("Error: could not send response back")
	}
}

func GrpcListener(appName string) {
	conn, err := grpc.NewClient("localhost:30000", grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	go httpRequestListenerLoop(stream, &wg)
	newHttpConnection(stream, appName)

	go func() {
		for response := range resDataChan {
			fmt.Printf("Response = \n%v\n", response)
			sendHttpresponse(stream, &response)
		}
	}()

	select {}
}
