package connhandler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	pb "github.com/Ahmed-Armaan/Localhost.git/proto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func tcpRequestListenerLoop(stream pb.TunnelService_TCPTunnelClient, wg *sync.WaitGroup) {
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
			if msg.GetMeta().ClientIp != "registered" {
				log.Fatal("Connection failed")
			} else {
				fmt.Printf("Connected: Access app at\n")
			}

		case pb.MessageType_DATA:
			fmt.Printf("Received request with connId%v\n:%v\n", msg.GetConnId(), msg)
			tcpResDataChan <- tcpreqData{
				connId: msg.GetConnId(),
				appId:  msg.GetAppId(),
				req:    msg,
			}

		case pb.MessageType_HEARTBEAT:
			fmt.Println("Heartbeat received")
		}
	}
}

func newTcpConnection(stream pb.TunnelService_TCPTunnelClient, appName string) {
	msg := &pb.TCPMessage{
		Type:  pb.MessageType_NEW_CONNECTION,
		AppId: appName,
	}
	if err := stream.Send(msg); err != nil {
		log.Println("failed to send NEW_CONNECTION:", err)
	}
}

func sendTcpResponse(stream pb.TunnelService_TCPTunnelClient, response *tcpreqData) {
	if err := stream.Send(response.req); err != nil {
		log.Println("Error: could not send response back")
	}
}

func GrpcTcpListener(appName string) {
	conn, err := grpc.NewClient("localhost:30000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTunnelServiceClient(conn)
	stream, err := client.TCPTunnel(context.Background())
	if err != nil {
		log.Fatalf("failed to start stream: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go tcpRequestListenerLoop(stream, &wg)
	newTcpConnection(stream, appName)

	// --- Persistent connection to local client ---
	go func() {
		localConn, err := net.Dial("tcp", "localhost:3001")
		if err != nil {
			log.Printf("Failed to connect to local test client: %v\n", err)
			return
		}
		defer localConn.Close()

		for response := range tcpResDataChan {
			fmt.Printf("Forwarding to local client: %v\n", response)

			data := response.req.GetData() // or response.Data, depending on your struct

			_, err := localConn.Write([]byte(data))
			if err != nil {
				log.Printf("Failed to send data to local client: %v\n", err)
				break
			}
		}
	}()

	select {}
}
