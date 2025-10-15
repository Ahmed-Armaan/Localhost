package httphandler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Listenhttp() {
	router := gin.Default()
	router.Any("/app/:appname/*apppath", portForwarder, reqhandler)

	err := router.Run(":9000")
	if err != nil {
		log.Fatal("Failed to start http Listener")
	}
}

func reqhandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Received and forwarder request",
	})
}
