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
				log.Printf("new HTTP connection from %s", msg.GetAppId())
				ActiveHttpConnmu.Lock()
				ActiveHttpConn[msg.GetAppId()] = TunnelConn{
					protocol:   int(Protocolhttp),
					httpStream: stream,
				}
				ActiveHttpConnmu.Unlock()

				connectionMessage := &pb.HTTPMessage{
					Type: pb.MessageType_NEW_CONNECTION,
					Payload: &pb.HTTPMessage_Response{
						Response: &pb.HTTPResponseData{
							StatusCode: 200,
							StatusText: `Connected: Access at URL "comingsoon.com"`,
						},
					},
				}

				if err := stream.Send(connectionMessage); err != nil {
					log.Println("Failed to send connection confirmation")
				}

			case pb.MessageType_CLOSE:
				ActiveHttpConnmu.Lock()
				delete(ActiveHttpConn, msg.GetConnId())
				ActiveHttpConnmu.Unlock()
				return

			case pb.MessageType_HEARTBEAT:
				pingBack := &pb.HTTPMessage{
					Type:    pb.MessageType_HEARTBEAT,
					RawData: []byte("pong"),
				}
				if err = stream.Send(pingBack); err != nil {
					log.Println("failed to send heartbeat:", err)
				}

			case pb.MessageType_ERROR:
				fmt.Printf("Error from App %s: %s\n", msg.GetAppId(), msg.GetErrorData())

			case pb.MessageType_DATA:
				res := msg.GetResponse()
				if res == nil {
					res = &pb.HTTPResponseData{
						StatusCode: 404,
						StatusText: "Not found",
					}
				}

				HttpResChansmu.Lock()
				if ch, ok := HttpResChans[msg.GetConnId()]; ok {
					ch <- res
				} else {
					log.Printf("Response channel not found for connId: %s", msg.GetConnId())
				}
				HttpResChansmu.Unlock()
			}
		}
	}()

	for {
		select {
		case req := <-HttpInreq:
			request := &pb.HTTPMessage{
				Type:   pb.MessageType_DATA,
				ConnId: req.connId,
				Payload: &pb.HTTPMessage_Request{
					Request: req.httpReq,
				},
			}

			ActiveHttpConnmu.RLock()
			conn, ok := ActiveHttpConn[req.appId]
			ActiveHttpConnmu.RUnlock()

			if !ok || conn.httpStream == nil {
				log.Printf("no active HTTP stream for appId: %s", req.appId)
				continue
			}

			if err := conn.httpStream.Send(request); err != nil {
				log.Printf("failed to send request to app %s: %v", req.appId, err)
				return err
			}

		case err := <-errChan:
			return err

		case <-stream.Context().Done():
			return nil
		}
	}
}

func (s *server) TCPTunnel(stream pb.TunnelService_TCPTunnelServer) error {
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
				log.Printf("new TCP connection from %s", msg.GetAppId())
				ActiveTcpConnmu.Lock()
				ActiveTcpConn[msg.GetAppId()] = TunnelConn{
					protocol:  int(Protocoltcp),
					tcpStream: stream,
				}
				ActiveTcpConnmu.Unlock()

				connMsg := &pb.TCPMessage{
					Type: pb.MessageType_NEW_CONNECTION,
					Meta: &pb.TCPReqData{
						ClientIp: "registered",
					},
				}
				if err := stream.Send(connMsg); err != nil {
					log.Printf("failed to send TCP registration confirmation: %v", err)
				}

			case pb.MessageType_CLOSE:
				ActiveTcpConnmu.Lock()
				delete(ActiveTcpConn, msg.GetAppId())
				ActiveTcpConnmu.Unlock()
				return

			case pb.MessageType_HEARTBEAT:
				pingBack := &pb.TCPMessage{
					Type: pb.MessageType_HEARTBEAT,
					Data: []byte("pong"),
				}
				if err = stream.Send(pingBack); err != nil {
					log.Println("failed to send TCP heartbeat:", err)
				}

			case pb.MessageType_ERROR:
				log.Printf("TCP error from app %s: %s", msg.GetAppId(), msg.GetErrorData())

			case pb.MessageType_DATA:
				TcpResChansmu.Lock()
				if ch, ok := TcpResChans[msg.GetConnId()]; ok {
					ch <- msg
				} else {
					log.Printf("TCP response channel not found for connId: %s", msg.GetConnId())
				}
				TcpResChansmu.Unlock()
			}
		}
	}()

	for {
		select {
		case req := <-TcpInreq:
			m := &pb.TCPMessage{
				Type:   pb.MessageType_DATA,
				ConnId: req.connId,
				Data:   req.tcpReq.GetData(),
				Meta:   req.tcpReq.GetMeta(),
			}

			ActiveTcpConnmu.RLock()
			conn, ok := ActiveTcpConn[req.appId]
			ActiveTcpConnmu.RUnlock()

			if !ok || conn.tcpStream == nil {
				log.Printf("no active TCP stream for appId: %s", req.appId)
				continue
			}

			if err := conn.tcpStream.Send(m); err != nil {
				log.Printf("failed to send TCP request to app %s: %v", req.appId, err)
				return err
			}

		case err := <-errChan:
			return err

		case <-stream.Context().Done():
			return nil
		}
	}
}
