package connhandler

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

var (
	tcpReqDataChan = make(chan []byte, 1024)
	tcpResDataChan = make(chan []byte, 1024)
)

func connectApp(port int) {
	appAddr := fmt.Sprintf("localhost:%d", port)
	conn, err := net.Dial("tcp", appAddr)
	if err != nil {
		log.Fatalf("Error: Could not connect to the app\n%v\n", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	go func() {
		for {
			response, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					log.Println("Connection closed by server")
				} else {
					log.Printf("Error reading from app: %v\n", err)
				}
				return
			} else {
				tcpResDataChan <- response
			}
		}
	}()

	for {
		payload := <-tcpReqDataChan
		if payload == nil {
			log.Println("Empty request received")
			continue
		}
		conn.Write(payload)
	}
}
