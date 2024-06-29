package xhttp

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewClient() *http.Client {
	h := http.DefaultClient
	h.Transport = otelhttp.NewTransport(h.Transport)
	return h
}
