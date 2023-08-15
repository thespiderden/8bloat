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
	"strconv"
	"strings"
	"sync"

	"github.com/bwmarrin/snowflake"
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

var Node *snowflake.Node

func init() {
	snowflake.Epoch = 1665230888000
}

func ShortID() string {
	idSlice := &bytes.Buffer{}
	binary.Write(idSlice, binary.LittleEndian, Node.Generate())
	return base64.RawURLEncoding.EncodeToString(idSlice.Bytes())
}

var closeLog func()

func init() {
	flag.Parse()

	if *writeConf && (*file != "") {
		log.Fatal("cannot use -f and -wc at the same time")
		os.Exit(1)
	}

	if *file == "-" {
		_, err := os.Stdout.Write(defaultConfig)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
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
	Node = nil
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
			if closeLog != nil {
				closeLog()
			}

			if val != "" {
				f, err := os.Open(val)
				if err != nil {
					return err
				}

				closeLog = func() { f.Close() }
				defer log.SetOutput(f)
			} else {
				log.SetOutput(os.Stdout)
				closeLog = nil
			}
		case "asset_stamp":
			if val == "snowflake" || val == "random" || val == "" {
				defer func() {
					AssetStamp = "." + ShortID()
				}()
			} else {
				AssetStamp = val
			}
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

			Node = node
		default:
			return errors.New("unknown config key " + key)
		}

		if Node == nil {
			var err error
			Node, err = snowflake.NewNode(0)
			if err != nil {
				panic(err)
			}
		}
	}

	return nil
}
