package connhandler

import (
	"fmt"
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
				log.Printf("new requuses from %s", msg.GetAppId())
				ActiveConnmu.Lock()
				ActiveConn[msg.GetAppId()] = TunnelConn{
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

			case pb.MessageType_ERROR:
				fmt.Printf("Error:%s\nfrom App %s\n", msg.GetErrorData(), msg.GetAppId())

			case pb.MessageType_DATA:
				res := msg.GetResponse()
				fmt.Printf("got message\n%v\n", msg)
				if res == nil {
					res = &pb.HTTPResponseData{
						StatusCode: 404,
						StatusText: "Not found",
					}
				}

				ResChansmu.Lock()
				if ch, ok := ResChans[msg.GetConnId()]; ok {
					fmt.Println("Written into chan")
					ch <- res
				} else {
					fmt.Println("No chan LOl")
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
