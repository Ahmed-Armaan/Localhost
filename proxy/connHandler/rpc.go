package connhandler

import (
	"log"

	pb "github.com/Ahmed-Armaan/Localhost.git/proto/proto"
)

type server struct {
	pb.UnimplementedTunnelServiceServer
}

func (s *server) HTTPTunnel(stream pb.TunnelService_HTTPTunnelServer) error {
	errChan := make(chan error)

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				errChan <- err
				return
			}

			switch msg.GetType() {
			case pb.MessageType_NEW_CONNECTION:
				log.Printf("new requuses from %s", msg.GetConnId())
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
				return

			case pb.MessageType_HEARTBEAT:
				pingBack := &pb.HTTPMessage{
					Type:    pb.MessageType_HEARTBEAT,
					RawData: []byte("pong"),
				}

				if err = stream.Send(pingBack); err != nil {
					log.Println("failed to send heartbeat: ", err)
				}

			case pb.MessageType_DATA:
				res := msg.GetResponse()
				if res == nil {
					res = &pb.HTTPResponseData{
						StatusCode: 404,
						StatusText: "Not found",
					}
				}

				ResChansmu.Lock()
				if ch, ok := ResChans[msg.GetConnId()]; ok {
					ch <- res
				}
				ResChansmu.Unlock()

			}
		}
	}()

	for {
		select {
		case req := <-Inreq:
			request := &pb.HTTPMessage{
				Type:   pb.MessageType_DATA,
				ConnId: req.connId,
				Payload: &pb.HTTPMessage_Request{
					Request: req.req,
				},
			}

			ActiveConnmu.Lock()
			if err := ActiveConn[req.appId].stream.Send(request); err != nil {
				return err
			}
			ActiveConnmu.Unlock()

		case err := <-errChan:
			return err
		case <-stream.Context().Done():
			return nil
		}
	}
}
