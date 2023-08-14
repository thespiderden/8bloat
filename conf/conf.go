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
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var lock sync.RWMutex

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

var (
	ListenAddress  string
	ClientName     string
	ClientScope    string
	ClientWebsite  string
	SingleInstance string
	PostFormats    []PostFormat
	LogFile        string
	AssetStamp     string
	SFNodeID       int
)

var DataFS fs.FS

//go:embed bloat.conf
var defaultConfig []byte

func init() {
	flag.Parse()

	if writeConf && file != "" {
		log.Fatal("cannot use -f and -wc at the same time")
		os.Exit(1)
	}

	if file == "-" {
		_, err := os.Stdout.Write(defaultConfig)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if file == "" {
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
				file = path
				break
			}
		}

		if file == "" {
			log.Fatal("exhausted default config search, please specify your own")
		}
	}

	file, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	err = readConf(file)
	if err != nil {
		log.Fatal("error parsing config:", err)
	}
}

func init() {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGHUP)

	go func() {
		for {
			<-sigch
			if file == "-" {
				log.Println("recieved sighup, but config is from stdin and cannot be reloaded")
				continue
			}

			f, err := os.Open(file)
			if err != nil {
				log.Println("recieved sighup, error while opening file:", err)
				continue
			}

			lock.RLock()

			err = readConf(f)
			f.Close()
			if err != nil {
				lock.RUnlock()
				log.Println("recieved sighup, error while parsing and applying config: ", err)
				continue
			}

			lock.RUnlock()
			log.Println("recieved sighup, reloaded config")
		}
	}()
}

func readConf(reader io.Reader) error {
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
			ListenAddress = val
		case "client_name":
			ClientName = val
		case "client_scope":
			ClientScope = val
		case "client_website":
			ClientWebsite = val
		case "single_instance":
			SingleInstance = val
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
			PostFormats = formats
		case "log_file":
			if val != "" {
				f, err := os.Open(val)
				if err != nil {
					return err
				}

				defer log.SetOutput(f)
			}
		case "asset_stamp":
			AssetStamp = val // Defer this to service.NewService
		case "snowflake_node_id":
			var no int
			if val != "" {
				var err error
				no, err = strconv.Atoi(val)
				if err != nil {
					log.Fatal("invalid config key: " + val)
				}
			}

			SFNodeID = no
		default:
			return errors.New("unknown config key " + key)
		}
	}

	return nil
}
