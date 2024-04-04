package main

import (
	"bufio"
	_ "embed"
	"errors"
	"io"
	"log"
	"spiderden.org/8bloat/internal/conf"
	"strconv"
	"strings"
	"time"
)

//go:embed bloat.conf
var DefaultConfig []byte

func readConf(reader io.Reader) (conf.Configuration, error) {
	var config conf.Configuration

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
			config.ListenAddress = val
		case "client_name":
			config.ClientName = val
		case "client_scope":
			config.ClientScope = val
		case "client_website":
			config.ClientWebsite = val
		case "single_instance":
			config.Instance = val
		case "user_agent":
			config.UserAgent = val
		case "database_path":
			// ignore
		case "post_formats":
			vals := strings.Split(val, ",")
			var formats []conf.PostFormat
			for _, v := range vals {
				pair := strings.Split(v, ":")
				if len(pair) != 2 {
					return config, errors.New("invalid config key " + key)
				}
				n := strings.TrimSpace(pair[0])
				t := strings.TrimSpace(pair[1])
				if len(n) < 1 || len(t) < 1 {
					return config, errors.New("invalid config key " + key)
				}
				formats = append(formats, conf.PostFormat{
					Name: n,
					Type: t,
				})
			}
			config.PostFormats = formats
		case "asset_stamp":
			config.AssetStamp = val
		case "snowflake_node_id":
			if val != "" {
				var err error
				config.Node, err = strconv.ParseInt(val, 10, 64)
				if err != nil || config.Node > 9 || config.Node < 0 {
					return config, errors.New("invalid config key: " + val)
				}
			}
		case "http_client_timeout":
			if val != "" {
				var err error

				config.RequestTimeout, err = time.ParseDuration(val)
				if err != nil {
					return config, err
				}
			}
		case "http_client_response_size_limit":
			if val != "" {
				i, err := strconv.Atoi(val)
				if err != nil {
					return config, errors.New("http_client_response_size_limit is not a number")
				}

				if i < 0 {
					return config, errors.New("http_client_response_size_limit cannot be negative")
				}

				config.ResponseLimit = int64(i)
			}
		default:
			return config, errors.New("unknown config key " + key)
		}
	}

	if config.UserAgent == "" {
		config.UserAgent = "8bloat/" + conf.Version() + " (Mastodon client, https://spiderden.org/projects/8bloat)"
	}

	if int64(config.RequestTimeout) == 0 {
		config.RequestTimeout = time.Second * 8
	}

	if config.ResponseLimit == 0 {
		config.ResponseLimit = (1 << (10 * 2)) * 8 // 8MB
	}

	return config, nil
}
