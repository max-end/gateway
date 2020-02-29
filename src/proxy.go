package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type ResponseTransport struct {
	StatusCode  int
	Target      string
	URL         string
	SuccessTime time.Time
	RequestTime time.Time
}

func (s *ResponseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	s.StatusCode = response.StatusCode
	return response, err
}
func doRequest(w *http.ResponseWriter, r *http.Request, target *url.URL, transport *ResponseTransport) {
	py := httputil.NewSingleHostReverseProxy(target)
	py.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
		ErrorHandler(writer, &Result{
			Status:  502,
			Message: "Bad Gateway",
		})
	}
	py.Transport = transport
	transport.Target = target.String() + r.URL.Path
	scheme := "http://"
	if r.TLS != nil {
		scheme = "https://"
	}
	transport.URL = scheme + r.Host + r.RequestURI
	py.ServeHTTP(*w, r)
	transport.SuccessTime = time.Now()
}
