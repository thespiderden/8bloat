package service

import (
	"net/http"
	"time"

	"spiderden.org/8b/conf"
	"spiderden.org/masta"
)

type tripper struct {
	underlying http.RoundTripper
}

func (t *tripper) RoundTrip(r *http.Request) (*http.Response, error) {
	response, err := t.underlying.RoundTrip(r)
	if response != nil && response.Body != nil {
		response.Body = http.MaxBytesReader(nil, response.Body, conf.Get().ResponseLimit)
	}

	return response, err
}

var client = http.Client{
	Transport: &tripper{
		underlying: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	},
	Timeout: 8 * time.Second,
}

func newMastaClient(cfg *masta.Config) *masta.Client {
	mclient := masta.NewClient(cfg)
	mclient.Client = client
	mclient.UserAgent = conf.Get().UserAgent
	return mclient
}
