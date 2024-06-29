package xhttp

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewClient() *http.Client {
	h := http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxConnsPerHost:       100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}

	h.Transport = otelhttp.NewTransport(h.Transport)
	return &h
}
