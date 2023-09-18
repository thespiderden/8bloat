// package conf handles user configuration at runtime, and stores
// module-global data.

package conf

import (
	"bufio"
	_ "embed"
	"errors"
	"flag"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/snowflake"
)

var version = "unknown"

func init() {
	snowflake.Epoch = 1665230888000
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

func Version() string {
	return version
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

	node int64
}

func (c Configuration) SingleInstance() (instance string, ok bool) {
	if len(c.Instance) > 0 {
		instance = c.Instance
		ok = true
	}
	return
}

var cfg atomic.Value

func Get() *Configuration {
	conf := cfg.Load()
	if conf == nil {
		return nil
	}

	cfg := conf.(Configuration)
	return &cfg
}

//go:embed bloat.conf
var defaultConfig []byte

func init() {
	flag.Parse()

	if *writeConf && (*file != "") {
		log.Fatal("cannot use -f and -wc at the same time")
		os.Exit(1)
	}

	if *writeConf {
		_, err := os.Stdout.Write(defaultConfig)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *file == "-" {
		if err := readConf(os.Stdin); err != nil {
			log.Fatal(err)
		}

		return
	}

	if *file == "" {
		var path string
		for _, path = range []string{"8bloat.conf", "/etc/8bloat.conf", "bloat.conf", "/etc/bloat.conf"} {
			stat, err := os.Stat(path)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				log.Fatal("error searching for config file: ", err)
			}

			if !stat.IsDir() {
				file = &path
				break
			}
		}

		if *file == "" {
			log.Fatal("exhausted default config search, please specify your own")
		}
	}

	file, err := os.Open(*file)
	if err != nil {
		log.Fatal(err)
	}

	err = readConf(file)
	if err != nil {
		log.Fatal("error parsing config:", err)
	}
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

func readConf(reader io.Reader) error {
	var conf Configuration

	scanner := bufio.NewScanner(reader)

	var nodeno int

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) < 1 {
			continue
		}

		index := strings.IndexRune(line, '#')
		if index == 0 {
			continue
		}

		index = strings.IndexRune(line, '=')
		if index < 1 {
			log.Fatal("invalid config key")
		}

		key := strings.TrimSpace(line[:index])
		val := strings.TrimSpace(line[index+1:])

		switch key {
		case "listen_address":
			conf.ListenAddress = val
		case "client_name":
			conf.ClientName = val
		case "client_scope":
			conf.ClientScope = val
		case "client_website":
			conf.ClientWebsite = val
		case "single_instance":
			conf.Instance = val
		case "user_agent":
			conf.UserAgent = val
		case "database_path":
			// ignore
		case "post_formats":
			vals := strings.Split(val, ",")
			var formats []PostFormat
			for _, v := range vals {
				pair := strings.Split(v, ":")
				if len(pair) != 2 {
					return errors.New("invalid config key " + key)
				}
				n := strings.TrimSpace(pair[0])
				t := strings.TrimSpace(pair[1])
				if len(n) < 1 || len(t) < 1 {
					return errors.New("invalid config key " + key)
				}
				formats = append(formats, PostFormat{
					Name: n,
					Type: t,
				})
			}
			conf.PostFormats = formats
		case "asset_stamp":
			conf.AssetStamp = val
		case "snowflake_node_id":
			if val != "" {
				var err error
				nodeno, err = strconv.Atoi(val)
				if err != nil {
					return errors.New("invalid config key: " + val)
				}
			}
		case "http_client_timeout":
			if val != "" {
				var err error

				conf.RequestTimeout, err = time.ParseDuration(val)
				if err != nil {
					return err
				}
			}
		case "http_client_response_size_limit":
			if val != "" {
				i, err := strconv.Atoi(val)
				if err != nil {
					return errors.New("http_client_response_size_limit is not a number")
				}

				if i < 0 {
					return errors.New("http_client_response_size_limit cannot be negative")
				}

				conf.ResponseLimit = int64(i)
			}
		default:
			return errors.New("unknown config key " + key)
		}
	}

	if conf.UserAgent == "" {
		conf.UserAgent = "8bloat/" + version + " (Mastodon client, https://spiderden.org/projects/8bloat)"
	}

	if int64(conf.RequestTimeout) == 0 {
		conf.RequestTimeout = time.Second * 8
	}

	if conf.ResponseLimit == 0 {
		conf.ResponseLimit = (1 << (10 * 2)) * 8 // 8MB
	}

	currcfg := Get()
	if node == nil || (currcfg != nil && currcfg.node != conf.node) {
		nodelock.Lock()

		nnode, err := snowflake.NewNode(int64(nodeno))
		if err != nil {
			nodelock.Unlock()
			return errors.New("error creating snowflake node: " + err.Error())
		}

		node = nnode
		nodelock.Unlock()
	}

	// random for backwards compatability
	if conf.AssetStamp == "snowflake" || conf.AssetStamp == "random" || conf.AssetStamp == "" {
		conf.AssetStamp = ID()
	}

	cfg.Store(conf)

	return nil
}
