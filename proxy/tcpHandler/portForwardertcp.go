package tcphandler

import (
	//	"context"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	//	"time"

	"github.com/Ahmed-Armaan/Localhost.git/proxy/connHandler"
	"github.com/google/uuid"

	pb "github.com/Ahmed-Armaan/Localhost.git/proto/proto"
)

type establishReq struct {
	App string `json:"app"`
}

func establishConn(conn net.Conn) (bool, error, chan *pb.TCPMessage, string, string) {
	// establish a new connection and map it to the app requested
	// read the JSON send as data packet
	reader := bufio.NewReader(conn)

	reqData, err := reader.ReadBytes('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read handshake: %w", err), nil, "", ""
	}

	var reqJson establishReq
	if err := json.Unmarshal(reqData, &reqJson); err != nil {
		conn.Write([]byte("Invalid payload\n"))
		return false, err, nil, "", ""
	}

	remoteAddr := strings.Split(conn.RemoteAddr().String(), ":")
	remotePort, err := strconv.Atoi(remoteAddr[1])
	if err != nil {
		return false, fmt.Errorf("invalid remote port: %w", err), nil, "", ""
	}

	currReq := &pb.TCPMessage{
		Type: pb.MessageType_NEW_CONNECTION,
		Meta: &pb.TCPReqData{
			TargetHost: reqJson.App,
			TargetPort: int32(remotePort),
			ClientIp:   remoteAddr[0],
		},
	}

	resChan := make(chan *pb.TCPMessage)
	connId := uuid.NewString()
	appName := reqJson.App
	connhandler.TcpRequestListener(currReq, resChan, connId, appName)
	return true, nil, resChan, connId, appName
}

func reqHandler(conn net.Conn, resChan chan *pb.TCPMessage, connId string, appName string) {
	defer conn.Close()

	go func() {
		defer func() {
			connhandler.TcpResponder(&pb.TCPMessage{
				Type: pb.MessageType_CLOSE,
			}, connId, appName)
		}()

		buff := make([]byte, 4096)
		for {
			n, err := conn.Read(buff)
			if err != nil {
				log.Printf("Connection closed or error: %v\n", err)
				return
			}
			if n > 0 {
				fmt.Printf("conn says: %s\n", buff[:n])

				dataCopy := make([]byte, n)
				copy(dataCopy, buff[:n])

				currReq := pb.TCPMessage{
					Type: pb.MessageType_DATA,
					Data: dataCopy,
				}
				connhandler.TcpResponder(&currReq, connId, appName)
			}
		}
	}()

	for res := range resChan {
		if res.Type == pb.MessageType_CLOSE {
			return
		}
		if len(res.Data) > 0 {
			_, err := conn.Write(res.Data)
			if err != nil {
				log.Printf("Could not write TCP payload: %v\n", err)
				return
			}
		}
	}
}
