package connhandler

import (
	"log"

	pb "github.com/Ahmed-Armaan/Localhost.git/proto/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pb.UnimplementedTunnelServiceServer
}

func (s *server) HTTPTunnel(stream pb.TunnelService_HTTPTunnelServer) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		switch msg.GetType() {
		case pb.MessageType_NEW_CONNECTION:
			ActiveConnmu.Lock()
			ActiveConn[msg.GetConnId()] = TunnelConn{
				protocol: int(Protocolhttp),
				stream:   stream,
			}
			ActiveConnmu.Unlock()

		case pb.MessageType_CLOSE:
			ActiveConnmu.Lock()
			delete(ActiveConn, msg.GetConnId())
			ActiveConnmu.Unlock()
			return nil

		case pb.MessageType_HEARTBEAT:
			pingBack := &pb.HTTPMessage{
				Type:    pb.MessageType_HEARTBEAT,
				RawData: []byte("pong"),
			}

			if err = stream.Send(pingBack); err != nil {
				log.Println("failed to send heartbeat: ", err)
			}
		}
	}
}
