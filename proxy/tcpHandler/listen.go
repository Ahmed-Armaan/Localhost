package tcphandler

import (
	"fmt"
	"log"
	"net"
)

func ListenTcp() {
	lis, err := net.Listen("tcp", "0.0.0.0:9001")
	if err != nil {
		log.Fatalf("Error: Can't start TCP server: %v\n", err)
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("Failed to accept TCP request: %v\n", err)
			continue
		}
		fmt.Printf("New request received:\n")
		go handleNewConn(conn)
	}
}

func handleNewConn(conn net.Conn) {
	established, err, resChan, connId, appName := establishConn(conn)
	if established {
		fmt.Println("Established")
		reqHandler(conn, resChan, connId, appName)
	} else {
		log.Printf("Rejected connection: %v\n", err)
		conn.Close()
	}
}
