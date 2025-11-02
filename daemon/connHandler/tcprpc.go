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
			if msg.GetMeta().GetClientIp() != "registered" {
				log.Println("Registration failed")
			} else {
				fmt.Println("Daemon registered with proxy")
			}

		case pb.MessageType_DATA:
			connId := msg.GetConnId()
			fmt.Printf("Received request for connId: %s\n", connId)

			tcpConnectionsMu.RLock()
			conn, exists := tcpConnections[connId]
			tcpConnectionsMu.RUnlock()

			if !exists {
				fmt.Printf("Creating new app connection for connId: %s\n", connId)
				conn = &tcpConnection{
					reqChan: make(chan []byte, 100),
					resChan: make(chan []byte, 100),
				}

				tcpConnectionsMu.Lock()
				tcpConnections[connId] = conn
				tcpConnectionsMu.Unlock()

				go connectApp(port, conn, connId)
				go sendTcpResponse(stream, conn.resChan, connId, appName)
			}

			if msg.GetData() != nil {
				conn.reqChan <- msg.GetData()
			}

		case pb.MessageType_CLOSE:
			connId := msg.GetConnId()
			fmt.Printf("Closing connection: %s\n", connId)

			tcpConnectionsMu.Lock()
			if conn, exists := tcpConnections[connId]; exists {
				close(conn.reqChan)
				if conn.appConn != nil {
					conn.appConn.Close()
				}
				delete(tcpConnections, connId)
			}
			tcpConnectionsMu.Unlock()
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

func sendTcpResponse(stream pb.TunnelService_TCPTunnelClient, resChan chan []byte, connId string, appName string) {
	for res := range resChan {
		msg := &pb.TCPMessage{
			Type:   pb.MessageType_DATA,
			Data:   res,
			ConnId: connId,
			AppId:  appName,
		}
		if err := stream.Send(msg); err != nil {
			log.Printf("Can't send response: %v\n", err)
			return
		}
		fmt.Printf("Sent response for connId: %s\n", connId)
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
