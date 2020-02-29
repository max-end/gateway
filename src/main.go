package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func ErrorHandler(writer http.ResponseWriter, result *Result) {
	buf, _ := json.Marshal(result)
	writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	writer.WriteHeader(result.Status)
	writer.Write(buf)
}

func main() {
	router := sync.Map{}
	go syncRouter(&router)
	http.HandleFunc("/favicon.ico", func(writer http.ResponseWriter, request *http.Request) {
	})
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		requestTime := time.Now()
		target, err := parseRouter(&router, request)
		if err != nil {
			ErrorHandler(writer, &Result{
				Status:  404,
				Message: "page not found",
			})
			return
		}
		transport := new(ResponseTransport)
		transport.RequestTime = requestTime
		doRequest(&writer, request, target, transport)
		defer func(transport *ResponseTransport) {
			go Logs(transport, request)
		}(transport)
	})
	http.ListenAndServe(":8000", nil)
}
