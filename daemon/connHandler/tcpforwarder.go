package connhandler

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type tcpConnection struct {
	reqChan chan []byte
	resChan chan []byte
	appConn net.Conn
}

var (
	tcpConnections   = make(map[string]*tcpConnection)
	tcpConnectionsMu sync.RWMutex
)

func connectApp(port int, conn *tcpConnection, connId string) {
	appAddr := fmt.Sprintf("localhost:%d", port)
	appConn, err := net.Dial("tcp", appAddr)
	if err != nil {
		log.Printf("Error: Could not connect to app for connId %s: %v\n", connId, err)
		close(conn.resChan)
		return
	}
	defer appConn.Close()

	conn.appConn = appConn
	fmt.Printf("Connected to app for connId: %s\n", connId)

	go func() {
		reader := bufio.NewReader(appConn)
		buff := make([]byte, 4096)
		for {
			n, err := reader.Read(buff)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading from app (connId %s): %v\n", connId, err)
				}
				close(conn.resChan)
				return
			}
			if n > 0 {
				data := make([]byte, n)
				copy(data, buff[:n])
				conn.resChan <- data
				fmt.Printf("Received %d bytes from app for connId: %s\n", n, connId)
			}
		}
	}()

	for payload := range conn.reqChan {
		if payload == nil {
			continue
		}
		_, err := appConn.Write(payload)
		if err != nil {
			log.Printf("Error writing to app (connId %s): %v\n", connId, err)
			return
		}
		fmt.Printf("Forwarded %d bytes to app for connId: %s\n", len(payload), connId)
	}
}
