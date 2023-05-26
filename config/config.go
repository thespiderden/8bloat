package config

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"

	"bloat/model"
)

type config struct {
	ListenAddress   string
	ClientName      string
	ClientScope     string
	ClientWebsite   string
	SingleInstance  string
	StaticDirectory string
	TemplatesPath   string
	CustomCSS       string
	PostFormats     []model.PostFormat
	LogFile         string
}

func (c *config) IsValid() bool {
	if len(c.ListenAddress) < 1 ||
		len(c.ClientName) < 1 ||
		len(c.ClientScope) < 1 ||
		len(c.ClientWebsite) < 1 ||
		len(c.StaticDirectory) < 1 ||
		len(c.TemplatesPath) < 1 {
		return false
	}
	return true
}

func Parse(r io.Reader) (c *config, err error) {
	c = new(config)
	scanner := bufio.NewScanner(r)
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
			return nil, errors.New("invalid config key")
		}

		key := strings.TrimSpace(line[:index])
		val := strings.TrimSpace(line[index+1 : len(line)])

		switch key {
		case "listen_address":
			c.ListenAddress = val
		case "client_name":
			c.ClientName = val
		case "client_scope":
			c.ClientScope = val
		case "client_website":
			c.ClientWebsite = val
		case "single_instance":
			c.SingleInstance = val
		case "static_directory":
			c.StaticDirectory = val
		case "templates_path":
			c.TemplatesPath = val
		case "custom_css":
			c.CustomCSS = val
		case "database_path":
			// ignore
		case "post_formats":
			vals := strings.Split(val, ",")
			var formats []model.PostFormat
			for _, v := range vals {
				pair := strings.Split(v, ":")
				if len(pair) != 2 {
					return nil, errors.New("invalid config key " + key)
				}
				n := strings.TrimSpace(pair[0])
				t := strings.TrimSpace(pair[1])
				if len(n) < 1 || len(t) < 1 {
					return nil, errors.New("invalid config key " + key)
				}
				formats = append(formats, model.PostFormat{
					Name: n,
					Type: t,
				})
			}
			c.PostFormats = formats
		case "log_file":
			c.LogFile = val
		default:
			return nil, errors.New("invalid config key " + key)
		}
	}

	return
}

func ParseFiles(files []string) (c *config, err error) {
	var lastErr error
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			lastErr = err
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			lastErr = err
			return nil, err
		}
		if info.IsDir() {
			continue
		}
		return Parse(f)
	}
	if lastErr == nil {
		lastErr = errors.New("invalid config file")
	}
	return nil, lastErr
}
