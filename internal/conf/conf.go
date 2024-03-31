// package conf handles user configuration at runtime, and stores
// module-global data.

package conf

import (
	_ "embed"
	"net/http"
	"runtime/debug"
	"time"
)

const SnowflakeEpoch = 1665230888000

var version = "unknown"

func Version() string {
	return version
}

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	var vers string
	var modified bool

	for _, v := range info.Settings {
		switch v.Key {
		case "vcs.revision":
			vers = v.Value
		case "vcs.modified":
			modified = (v.Value == "true")
		}
	}

	if vers == "" {
		return
	}

	version = vers

	if modified {
		version += "*"
	}
}

type PostFormat struct {
	Name string
	Type string
}

type Configuration struct {
	ListenAddress  string
	ClientName     string
	ClientScope    string
	ClientWebsite  string
	PostFormats    []PostFormat
	AssetStamp     string
	UserAgent      string
	Instance       string
	ResponseLimit  int64
	RequestTimeout time.Duration
	Node           int64
}

func (c Configuration) SingleInstance() (instance string, ok bool) {
	if c.Instance != "" {
		instance, ok = c.Instance, true
	}
	return
}

func init() {
	http.DefaultTransport = &http.Transport{
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       3 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	http.DefaultClient = &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   8 * time.Second,
	}
}
