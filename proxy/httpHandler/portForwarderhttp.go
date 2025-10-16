package httphandler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	pb "github.com/Ahmed-Armaan/Localhost.git/proto/proto"
	"github.com/Ahmed-Armaan/Localhost.git/proxy/connHandler"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func abortReq(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": message,
	})
	c.Abort()
}

func portForwarder(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	method := c.Request.Method
	body := c.Request.Body
	defer body.Close()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		abortReq(c, http.StatusBadGateway, "can not read the request body")
		return
	}
	if len(bodyBytes) == 0 && ((method == "POST") || (method == "PUT") || (method == "PATCH")) {
		abortReq(c, http.StatusBadRequest, fmt.Sprintf("request body required for %s requests", method))
		return
	}

	header := make(map[string]*pb.HeaderValues)
	for k, v := range c.Request.Header {
		header[k] = &pb.HeaderValues{Values: v}
	}

	currReq := pb.HTTPRequestData{
		Method:  method,
		Path:    c.Param("apppath"),
		Query:   c.Request.URL.RawQuery,
		Headers: header,
		Body:    bodyBytes,
	}
	resChan := make(chan *pb.HTTPResponseData)

	connhandler.RequestListener(&currReq, resChan, uuid.NewString(), c.Param("appname"))

	select {
	case res := <-resChan:
		fmt.Printf("received from chan\n%v\n", res)
		if res == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Not found",
			})
			return
		}

		for k, v := range res.Headers {
			for _, val := range v.Values {
				c.Header(k, val)
			}
		}

		contentType := "application/octet-stream"
		if ct, ok := res.Headers["Content-Type"]; ok && len(ct.Values) > 0 {
			contentType = ct.Values[0]
		}
		c.Data(int(res.StatusCode), contentType, res.Body)

	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error": "Request timed out",
			})
		} else {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error": "Request cancelled",
			})
		}
	}

	fmt.Println(string(bodyBytes))
	c.Next()
}
