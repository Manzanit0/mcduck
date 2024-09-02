package xhttp

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewClient() *http.Client {
	// FIXME: these large timeouts are because some services take a long time in
	// their requests; these shouldn't be the defaults for every service though.
	h := http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			MaxConnsPerHost:       100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       120 * time.Second,
			ResponseHeaderTimeout: 120 * time.Second,
		},
	}

	h.Transport = otelhttp.NewTransport(h.Transport)
	return &h
}
