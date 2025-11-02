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

func tcpListenerLoop(stream pb.TunnelService_TCPTunnelClient, wg *sync.WaitGroup, port int, appName string) {
	defer wg.Done()

	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("Stream closed")
			} else {
				log.Printf("Stream error: %v\n", err)
			}
			return
		}

		switch msg.GetType() {
		case pb.MessageType_NEW_CONNECTION:
			if msg.GetMeta().ClientIp != "registered" {
				log.Fatalf("Connection failed")
			} else {
				fmt.Println("Connected")
				connId := msg.GetConnId()
				go connectApp(port)
				go sendTcpResponse(stream, connId, appName)
			}

		case pb.MessageType_DATA:
			fmt.Printf("Received request with connId%v\n:%v\n", msg.GetConnId(), msg)
			if msg.GetData() != nil {
				tcpReqDataChan <- msg.GetData()
			}
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

func sendTcpResponse(stream pb.TunnelService_TCPTunnelClient, connId string, appName string) {
	for res := range tcpResDataChan {
		msg := &pb.TCPMessage{
			Type:   pb.MessageType_DATA,
			Data:   res,
			ConnId: connId,
			AppId:  appName,
		}

		if err := stream.Send(msg); err != nil {
			log.Printf("Cant send response: %v\n", err)
			return
		}
	}
}

func GrpcTcpListener(appName string, port int) {
	conn, err := grpc.NewClient("localhost:30000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer conn.Close()

	client := pb.NewTunnelServiceClient(conn)
	stream, err := client.TCPTunnel(context.Background())
	if err != nil {
		log.Fatalf("failed to start stream: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	newTcpConnection(stream, appName)
	go tcpListenerLoop(stream, &wg, port, appName)
	wg.Wait()
}
