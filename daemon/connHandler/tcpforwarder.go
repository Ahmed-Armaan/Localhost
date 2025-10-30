package connhandler

import (
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/Ahmed-Armaan/Localhost.git/proto/proto"
)

type tcpreqData struct {
	connId string
	appId  string
	req    *pb.TCPMessage
}

var (
	tcpReqDataChan = make(chan tcpreqData, 1024)
	tcpResDataChan = make(chan tcpreqData, 1024)
)

func TcpReqForwarder(port int) {
	for {
		incoming := <-tcpReqDataChan
		if incoming.req == nil || incoming.req.Meta == nil {
			log.Println("TcpReqForwarder: invalid request (nil meta)")
			continue
		}

		meta := incoming.req.Meta
		target := fmt.Sprintf("%s:%d", meta.TargetHost, meta.TargetPort)
		conn, err := net.Dial("tcp", target)
		if err != nil {
			log.Printf("TcpReqForwarder: failed to connect to %s: %v", target, err)
			sendTCPError(incoming, fmt.Sprintf("connect error: %v", err))
			continue
		}

		// Write the initial data if present
		if len(incoming.req.Data) > 0 {
			if _, err := conn.Write(incoming.req.Data); err != nil {
				log.Printf("TcpReqForwarder: write failed: %v", err)
				conn.Close()
				sendTCPError(incoming, fmt.Sprintf("write error: %v", err))
				continue
			}
		}

		// Start goroutine to read responses from target service
		go func(localConn net.Conn, req tcpreqData) {
			defer localConn.Close()

			buf := make([]byte, 4096)
			for {
				n, err := localConn.Read(buf)
				if err != nil {
					if err != io.EOF {
						log.Printf("TcpReqForwarder: read error: %v", err)
					}
					// Send CLOSE signal
					tcpResDataChan <- tcpreqData{
						connId: req.connId,
						appId:  req.appId,
						req: &pb.TCPMessage{
							Type:   pb.MessageType_CLOSE,
							ConnId: req.connId,
							AppId:  req.appId,
						},
					}
					return
				}

				// Send data chunk back through tunnel
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				tcpResDataChan <- tcpreqData{
					connId: req.connId,
					appId:  req.appId,
					req: &pb.TCPMessage{
						Type:   pb.MessageType_DATA,
						ConnId: req.connId,
						AppId:  req.appId,
						Data:   chunk,
					},
				}
			}
		}(conn, incoming)
	}
}

func sendTCPError(req tcpreqData, msg string) {
	tcpResDataChan <- tcpreqData{
		connId: req.connId,
		appId:  req.appId,
		req: &pb.TCPMessage{
			Type:      pb.MessageType_ERROR,
			ConnId:    req.connId,
			AppId:     req.appId,
			ErrorData: msg,
		},
	}
}
