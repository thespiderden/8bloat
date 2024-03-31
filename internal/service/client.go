package service

import (
	"net/http"
	"spiderden.org/8b/internal/conf"
	"time"
)

type tripper struct {
	underlying http.RoundTripper
	conf       conf.Configuration
}

func (t *tripper) RoundTrip(r *http.Request) (*http.Response, error) {
	response, err := t.underlying.RoundTrip(r)
	if response != nil && response.Body != nil {
		response.Body = http.MaxBytesReader(nil, response.Body, t.conf.ResponseLimit)
	}

	return response, err
}

func newClient(config conf.Configuration) *http.Client {
	return &http.Client{
		Transport: &tripper{
			conf: config,
			underlying: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}
