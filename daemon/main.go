package main

import (
	"log"
	"os"
	"strconv"

	"github.com/Ahmed-Armaan/Localhost.git/daemon/connHandler"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Expected os args\nrun <appName> <port>\n")
	}
	appName := os.Args[1]
	appPort := os.Args[2]

	port, err := strconv.Atoi(appPort)
	if err != nil {
		log.Fatalf("appPort expected to be an integer\n")
	}

	connhandler.GrpcListener(appName)
	connhandler.ReqForwarder(port)
}
