package connhandler

import (
	"fmt"
	"log"
	"net"
	"net/http"
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
	stream   any
}

var ActiveConn = make(map[string]TunnelConn)
var ActiveConnmu sync.RWMutex

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

func requestListener() {
	PORT := 8081
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil); err != nil {
		log.Fatal("Cant start request Listener")
	}
}

func listen() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		grpcListener()
	}()

	go func() {
		defer wg.Done()
		requestListener()
	}()

	wg.Wait()
}
