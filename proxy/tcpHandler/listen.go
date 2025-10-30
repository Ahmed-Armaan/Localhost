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
	fmt.Println("1")

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("Failed to accept TCP request: %v\n", err)
			continue
		}

		go handleNewConn(conn)
	}
}

func handleNewConn(conn net.Conn) {
	established, err, resChan := establishConn(conn)
	if established {
		reqHandler(conn, resChan)
	} else {
		log.Printf("Rejected connection: %v\n", err)
		conn.Close()
	}
}
