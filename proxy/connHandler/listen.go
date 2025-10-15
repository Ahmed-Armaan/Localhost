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
	protocol int
	stream   pb.TunnelService_HTTPTunnelServer
}

type tunnelWriter struct {
	req    *pb.HTTPRequestData
	appId  string
	connId string
}

var (
	ActiveConn   = make(map[string]TunnelConn)
	ActiveConnmu sync.RWMutex
	Inreq        = make(chan *tunnelWriter, 2048)
	ResChans     = make(map[string]chan *pb.HTTPResponseData)
	ResChansmu   sync.RWMutex
)

func grpcListener() {
	PORT := 8080
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterTunnelServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to start grpc server: %v", err)
	}
}

func RequestListener(request *pb.HTTPRequestData, resChan chan *pb.HTTPResponseData, connId string, appId string) {
	ResChansmu.Lock()
	ResChans[connId] = resChan
	ResChansmu.Unlock()

	Inreq <- &tunnelWriter{
		req:    request,
		appId:  appId,
		connId: connId,
	}
}

func Listen() {
	grpcListener()
}
