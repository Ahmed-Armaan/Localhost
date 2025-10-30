package main

import (
	"fmt"
	"sync"

	"github.com/Ahmed-Armaan/Localhost.git/proxy/connHandler"
	"github.com/Ahmed-Armaan/Localhost.git/proxy/httpHandler"
	"github.com/Ahmed-Armaan/Localhost.git/proxy/tcpHandler"
)

func main() {
	var wg sync.WaitGroup

	wg.Go(func() {
		httphandler.Listenhttp()
	})

	wg.Go(func() {
		tcphandler.ListenTcp()
	})

	wg.Go(func() {
		connhandler.Listen()
	})

	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		httphandler.Listenhttp()
	//	}()
	//
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		connhandler.Listen()
	//	}()

	fmt.Println("Server running")
	wg.Wait()
}
