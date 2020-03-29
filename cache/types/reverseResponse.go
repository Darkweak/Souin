package types

import (
	"net/http"
	"net/http/httputil"
)

// ReverseResponse object contains the response from reverse-proxy
type ReverseResponse struct {
	Response string
	Proxy    *httputil.ReverseProxy
	Request  *http.Request
}
