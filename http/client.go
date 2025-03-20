package http

import (
	"net/http"
	"time"
)

func NewHttpClient() *http.Client {
	transport := NewTransport()

	return &http.Client{
		Transport: transport,
		Timeout:   time.Duration(1) * time.Second,
	}
}

func NewTransport() http.RoundTripper {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DisableKeepAlives = true
	t.MaxConnsPerHost = 1
	//t.IdleConnTimeout = 30 * time.Second
	//t.MaxIdleConns = 1
	//t.MaxIdleConnsPerHost = 1
	return t
}
