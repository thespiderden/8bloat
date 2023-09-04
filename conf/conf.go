// package conf handles user configuration at runtime, and stores
// module-global data.

package conf

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"io"
	"io/fs"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/bwmarrin/snowflake"
)

var lock sync.RWMutex

var version string = "unknown"

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

func Lock() {
	lock.Lock()
}

func Unlock() {
	lock.Unlock()
}

type PostFormat struct {
	Name string
	Type string
}

type Configuration struct {
	ListenAddress string
	ClientName    string
	ClientScope   string
	ClientWebsite string
	PostFormats   []PostFormat
	AssetStamp    string
	Node          *snowflake.Node
	UserAgent     string

	Instance string
}

func (c Configuration) SingleInstance() (instance string, ok bool) {
	if len(c.Instance) > 0 {
		instance = c.Instance
		ok = true
	}
	return
}

func (c Configuration) shortID() string {
	idSlice := &bytes.Buffer{}
	binary.Write(idSlice, binary.LittleEndian, c.Node.Generate())
	return base64.RawURLEncoding.EncodeToString(idSlice.Bytes())
}

func ShortID() string {
	return Get().shortID()
}

func init() {
	snowflake.Epoch = 1665230888000
}

var cfg atomic.Value

func storeConf(c Configuration) {
	cfg.Store(c)
}

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

func readConf(reader io.Reader) error {
	var conf Configuration

	scanner := bufio.NewScanner(reader)
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
			var no int
			if val != "" {
				var err error
				no, err = strconv.Atoi(val)
				if err != nil {
					return errors.New("invalid config key: " + val)
				}
			}

			node, err := snowflake.NewNode(int64(no))
			if err != nil {
				return err
			}

			conf.Node = node
		default:
			return errors.New("unknown config key " + key)
		}
	}

	if conf.Node == nil {
		var err error
		conf.Node, err = snowflake.NewNode(0)
		if err != nil {
			panic(err)
		}
	}

	if conf.UserAgent == "" {
		conf.UserAgent = "8bloat/" + version + " (Mastodon client, https://spiderden.org/projects/8bloat)"
	}

	// random for backwards compatability
	if conf.AssetStamp == "snowflake" || conf.AssetStamp == "random" || conf.AssetStamp == "" {
		conf.AssetStamp = "." + conf.shortID()
	}

	storeConf(conf)
	return nil
}
