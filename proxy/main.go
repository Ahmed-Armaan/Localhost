package main

import (
	"fmt"
	"sync"

	"github.com/Ahmed-Armaan/Localhost.git/proxy/connHandler"
	"github.com/Ahmed-Armaan/Localhost.git/proxy/httpHandler"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		httphandler.Listenhttp()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		connhandler.Listen()
	}()

	fmt.Println("Server running")
	wg.Wait()
}
