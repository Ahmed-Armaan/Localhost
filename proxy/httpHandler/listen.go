package httphandler

import (
	"log"
	"net/http"
)

func Listenhttp() {
	http.HandleFunc("/app/{appname}/{apppath}", portForwarder)

	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal("Could not start server")
	}
}
