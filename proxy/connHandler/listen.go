package connhandler

import (
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/Ahmed-Armaan/Localhost.git/proto/proto"
	"google.golang.org/grpc"
)

type Protocol int

const (
	Protocolhttp Protocol = iota
	Protocoltcp
)

type TunnelConn struct {
	protocol   int
	httpStream pb.TunnelService_HTTPTunnelServer
	tcpStream  pb.TunnelService_TCPTunnelServer
}

type tunnelWriter struct {
	httpReq *pb.HTTPRequestData
	tcpReq  *pb.TCPMessage
	appId   string
	connId  string
}

var (
	ActiveHttpConn   = make(map[string]TunnelConn)
	ActiveHttpConnmu sync.RWMutex
	HttpInreq        = make(chan *tunnelWriter, 2048)
	HttpResChans     = make(map[string]chan *pb.HTTPResponseData)
	HttpResChansmu   sync.RWMutex
)

var (
	ActiveTcpConn   = make(map[string]TunnelConn)
	ActiveTcpConnmu sync.RWMutex
	TcpInreq        = make(chan *tunnelWriter, 2048)
	TcpResChans     = make(map[string]chan *pb.TCPMessage)
	TcpResChansmu   sync.RWMutex
)

func grpcListener() {
	PORT := 8080
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterTunnelServiceServer(s, &server{})
	log.Printf("gRPC tunnel server listening on port %d", PORT)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to start grpc server: %v", err)
	}
}

func HttpRequestListener(request *pb.HTTPRequestData, resChan chan *pb.HTTPResponseData, connId string, appId string) {
	HttpResChansmu.Lock()
	HttpResChans[connId] = resChan
	HttpResChansmu.Unlock()

	HttpInreq <- &tunnelWriter{
		httpReq: request,
		appId:   appId,
		connId:  connId,
	}
}

func TcpRequestListener(request *pb.TCPMessage, resChan chan *pb.TCPMessage, connId string, appId string) {
	TcpResChansmu.Lock()
	TcpResChans[connId] = resChan
	TcpResChansmu.Unlock()

	TcpInreq <- &tunnelWriter{
		tcpReq: request,
		appId:  appId,
		connId: connId,
	}
}

func Listen() {
	grpcListener()
}
