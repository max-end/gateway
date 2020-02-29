package main

import (
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
	"log"
	"net/http"
	"strings"
	"time"
)

type Request struct {
	Address     string            `json:"address"`
	Url         string            `json:"url"`
	Target      string            `json:"target"`
	Method      string            `json:"method"`
	Header      map[string]string `json:"header"`
	StatusCode  int               `json:"status_code"`
	RequestTime time.Time         `json:"request_time"`
	Time        int64             `json:"time"`
}

var esClient *elasticsearch.Client

func init() {
	var err error
	esClient, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://docker:9200"},
	})
	if err != nil {
		log.Panic(err)
	}
	log.Println("elastic search connect success")
}
func Logs(transport *ResponseTransport, request *http.Request) {
	body := Parse(transport, request)
	b, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
		return
	}
	req := esapi.IndexRequest{
		Index:   "request",
		Body:    strings.NewReader(string(b)),
		Refresh: "true",
	}
	_, err = req.Do(context.Background(), esClient)
	if err != nil {
		log.Println(err)
	}
}
func Parse(transport *ResponseTransport, request *http.Request) Request {
	var body Request
	header := make(map[string]string)
	for i := range request.Header {
		header[i] = request.Header.Get(i)
	}
	body.Header = header
	body.RequestTime = transport.RequestTime
	body.Time = transport.SuccessTime.UnixNano() - transport.RequestTime.UnixNano()
	body.Address = request.RemoteAddr
	body.Target = transport.Target
	body.Url = transport.URL
	body.StatusCode = transport.StatusCode
	body.Method = request.Method
	return body
}
