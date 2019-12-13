package config

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

type config struct {
	ListenAddress        string
	ClientName           string
	ClientScope          string
	ClientWebsite        string
	StaticDirectory      string
	TemplatesGlobPattern string
	DatabasePath         string
	Logfile              string
}

func (c *config) IsValid() bool {
	if len(c.ListenAddress) < 1 ||
		len(c.ClientName) < 1 ||
		len(c.ClientScope) < 1 ||
		len(c.ClientWebsite) < 1 ||
		len(c.StaticDirectory) < 1 ||
		len(c.TemplatesGlobPattern) < 1 ||
		len(c.DatabasePath) < 1 {
		return false
	}
	return true
}

func getDefaultConfig() *config {
	return &config{
		ListenAddress:        ":8080",
		ClientName:           "web",
		ClientScope:          "read write follow",
		ClientWebsite:        "http://localhost:8080",
		StaticDirectory:      "static",
		TemplatesGlobPattern: "templates/*",
		DatabasePath:         "database.db",
		Logfile:              "",
	}
}

func Parse(r io.Reader) (c *config, err error) {
	c = getDefaultConfig()
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
		case "static_directory":
			c.StaticDirectory = val
		case "templates_glob_pattern":
			c.TemplatesGlobPattern = val
		case "database_path":
			c.DatabasePath = val
		case "logfile":
			c.Logfile = val
		default:
			return nil, errors.New("invliad config key " + key)
		}
	}

	return
}

func ParseFile(file string) (c *config, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return
	}

	if info.IsDir() {
		return nil, errors.New("invalid config file")
	}

	return Parse(f)
}
