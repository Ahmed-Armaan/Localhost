package connhandler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	pb "github.com/Ahmed-Armaan/Localhost.git/proto/proto"
)

type reqData struct {
	connId string
	appId  string
	req    *pb.HTTPMessage
}

var (
	reqDataChan = make(chan reqData, 1024)
	resDataChan = make(chan reqData, 1024)
)

func buildHTTPHeader(req *pb.HTTPRequestData) http.Header {
	headers := http.Header{}
	for key, vals := range req.Headers {
		for _, v := range vals.Values {
			headers.Add(key, v)
		}
	}
	return headers
}

func HttpReqForwarder(port int) {
	for {
		req := <-reqDataChan
		if req.req == nil || req.req.GetRequest() == nil {
			log.Println("nil request received")
			continue
		}

		request := req.req.GetRequest()
		url := fmt.Sprintf("http://localhost:%d%s", port, request.GetPath())
		if request.Query != "" {
			url += "?" + request.Query
		}

		httpReq, err := http.NewRequest(
			request.Method,
			url,
			bytes.NewReader(request.Body),
		)
		if err != nil {
			log.Println("failed to create request:", err)
			continue
		}
		httpReq.Header = buildHTTPHeader(request)

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			log.Println("HTTP forward error:", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("failed to read response:", err)
			continue
		}

		headers := map[string]*pb.HeaderValues{}
		for k, v := range resp.Header {
			headers[k] = &pb.HeaderValues{Values: v}
		}

		res := &pb.HTTPResponseData{
			StatusCode: int32(resp.StatusCode),
			StatusText: resp.Status,
			Headers:    headers,
			Body:       body,
		}

		resDataChan <- reqData{
			connId: req.connId,
			appId:  req.appId,
			req: &pb.HTTPMessage{
				Type:    pb.MessageType_DATA,
				ConnId:  req.connId,
				AppId:   req.appId,
				Payload: &pb.HTTPMessage_Response{Response: res},
			},
		}

		resp.Body.Close()
	}
}
