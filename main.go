package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"bloat/config"
	"bloat/renderer"
	"bloat/service"
)

//go:embed templates/* static/*
var embedFS embed.FS

var defaultConfigs = []string{"bloat.conf", "/etc/bloat.conf"}

func errExit(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

func main() {
	configFile := flag.String("f", "", `config file, use a dash for stdin`)
	flag.Parse()

	var conf *config.Config
	var err error

	switch *configFile {
	case "-":
		conf, err = config.Parse(os.Stdin)
		if err != nil {
			errExit(err)
		}
	default:
		conf, err = config.ParseFiles(defaultConfigs)
		if err != nil {
			errExit(err)
		}
	}

	if !conf.IsValid() {
		errExit(errors.New("invalid config"))
	}

	templatesGlobPattern := "templates/*"
	renderer, err := renderer.NewRenderer(templatesGlobPattern, embedFS)
	if err != nil {
		errExit(err)
	}

	var logger *log.Logger
	if len(conf.LogFile) < 1 {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		lf, err := os.OpenFile(conf.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			errExit(err)
		}
		defer lf.Close()
		logger = log.New(lf, "", log.LstdFlags)
	}

	s := service.NewService(conf.ClientName, conf.ClientScope,
		conf.ClientWebsite, conf.SingleInstance,
		conf.PostFormats, renderer)
	handler := service.NewHandler(s, logger, embedFS)

	logger.Println("listening on", conf.ListenAddress)
	err = http.ListenAndServe(conf.ListenAddress, handler)
	if err != nil {
		errExit(err)
	}
}
