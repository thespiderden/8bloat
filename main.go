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

var (
	configFiles = []string{"bloat.conf", "/etc/bloat.conf"}
)

func errExit(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

func main() {
	configFile := flag.String("f", "", "config file")
	flag.Parse()

	if len(*configFile) > 0 {
		configFiles = []string{*configFile}
	}
	config, err := config.ParseFiles(configFiles)
	if err != nil {
		errExit(err)
	}

	if !config.IsValid() {
		errExit(errors.New("invalid config"))
	}

	templatesGlobPattern := "templates/*"
	renderer, err := renderer.NewRenderer(templatesGlobPattern, embedFS)
	if err != nil {
		errExit(err)
	}

	var logger *log.Logger
	if len(config.LogFile) < 1 {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		lf, err := os.OpenFile(config.LogFile,
			os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			errExit(err)
		}
		defer lf.Close()
		logger = log.New(lf, "", log.LstdFlags)
	}

	s := service.NewService(config.ClientName, config.ClientScope,
		config.ClientWebsite, config.SingleInstance,
		config.PostFormats, renderer)
	handler := service.NewHandler(s, logger, embedFS)

	logger.Println("listening on", config.ListenAddress)
	err = http.ListenAndServe(config.ListenAddress, handler)
	if err != nil {
		errExit(err)
	}
}
