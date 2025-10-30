package tcphandler

import (
	//	"context"
	"bufio"
	"encoding/json"
	"fmt"
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

func establishConn(conn net.Conn) (bool, error, chan *pb.TCPMessage) {
	// establish a new connection and map it to the app requested
	// read the JSON send as data packet
	reader := bufio.NewReader(conn)

	reqData, err := reader.ReadBytes('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read handshake: %w", err), nil
	}
	fmt.Println(reader)

	var reqJson establishReq
	if err := json.Unmarshal(reqData, &reqJson); err != nil {
		conn.Write([]byte("Invalid payload\n"))
		return false, err, nil
	}

	remoteAddr := strings.Split(conn.RemoteAddr().String(), ":")
	remotePort, err := strconv.Atoi(remoteAddr[1])
	if err != nil {
		return false, fmt.Errorf("invalid remote port: %w", err), nil
	}

	currReq := &pb.TCPMessage{
		Type: pb.MessageType_NEW_CONNECTION,
		Meta: &pb.TCPReqData{
			TargetHost: reqJson.App, // "App" used to identify which local app
			TargetPort: int32(remotePort),
			ClientIp:   remoteAddr[0],
		},
	}

	resChan := make(chan *pb.TCPMessage)
	connhandler.TcpRequestListener(currReq, resChan, uuid.NewString(), reqJson.App)
	return true, nil, resChan
}

func reqHandler(conn net.Conn, resChan chan *pb.TCPMessage) {
	for res := range resChan {
		if res.Type == pb.MessageType_CLOSE {
			conn.Close()
			return
		}
		if len(res.Data) > 0 {
			conn.Write(res.Data)
		}
	}
}
