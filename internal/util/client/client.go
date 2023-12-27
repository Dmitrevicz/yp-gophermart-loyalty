// Package client contains custom wrapper for http.Client.
package client

import (
	"net/http"
	"time"
)

// DefaultHTTPClientTimeoutSeconds - custom default http client timeout in seconds.
const DefaultHTTPClientTimeoutSeconds = 10

// NewClientDefault returns *http.Client with timeout set as DefaultHTTPClientTimeoutSeconds.
func NewClientDefault() *http.Client {
	return &http.Client{
		Timeout: time.Second * DefaultHTTPClientTimeoutSeconds,
	}
}
