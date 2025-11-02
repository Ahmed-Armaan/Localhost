package main

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/Ahmed-Armaan/Localhost.git/daemon/connHandler"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatalf("Expected os args\nrun <appName> <port>\n")
	}
	appName := os.Args[1]
	appPort := os.Args[2]
	protocol := os.Args[3]

	port, err := strconv.Atoi(appPort)
	if err != nil {
		log.Fatalf("appPort expected to be an integer\n")
	}

	switch protocol {
	case "http":
		go connhandler.GrpcListener(appName)
		connhandler.HttpReqForwarder(port)
	case "tcp":
		var wg sync.WaitGroup
		wg.Go(func() {
			connhandler.GrpcTcpListener(appName, port)
		})
		wg.Wait()
	}
}
