package grequest

import (
	"net/http"
	"time"
)

var _DefaultClient *http.Client

func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{Transport: http.DefaultTransport, Timeout: timeout}
}

func SetDefaultClient(client *http.Client) {
	_DefaultClient = client
}

func DefaultClient() *http.Client {
	return _DefaultClient
}
