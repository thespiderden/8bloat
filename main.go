package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bloat/config"
	"bloat/renderer"
	"bloat/repo"
	"bloat/service"
	"bloat/util"
)

var (
	configFile = "/etc/bloat.conf"
)

func errExit(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

func setupHttp() {
	tr := http.DefaultTransport.(*http.Transport)
	tr.MaxIdleConnsPerHost = 30
	tr.MaxIdleConns = 300
	tr.ForceAttemptHTTP2 = false
	tr.DialContext = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 3 * time.Minute,
		DualStack: true,
	}).DialContext
	client := http.DefaultClient
	client.Transport = tr
}

func main() {
	opts, _, err := util.Getopts(os.Args, "f:")
	if err != nil {
		errExit(err)
	}

	for _, opt := range opts {
		switch opt.Option {
		case 'f':
			configFile = opt.Value
		}
	}

	config, err := config.ParseFile(configFile)
	if err != nil {
		errExit(err)
	}

	if !config.IsValid() {
		errExit(errors.New("invalid config"))
	}

	templatesGlobPattern := filepath.Join(config.TemplatesPath, "*")
	renderer, err := renderer.NewRenderer(templatesGlobPattern)
	if err != nil {
		errExit(err)
	}

	err = os.Mkdir(config.DatabasePath, 0755)
	if err != nil && !os.IsExist(err) {
		errExit(err)
	}

	sessionDBPath := filepath.Join(config.DatabasePath, "session")
	sessionDB, err := util.NewDatabse(sessionDBPath)
	if err != nil {
		errExit(err)
	}

	appDBPath := filepath.Join(config.DatabasePath, "app")
	appDB, err := util.NewDatabse(appDBPath)
	if err != nil {
		errExit(err)
	}

	sessionRepo := repo.NewSessionRepo(sessionDB)
	appRepo := repo.NewAppRepo(appDB)

	customCSS := config.CustomCSS
	if len(customCSS) > 0 && !strings.HasPrefix(customCSS, "http://") &&
		!strings.HasPrefix(customCSS, "https://") {
		customCSS = "/static/" + customCSS
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

	setupHttp()

	s := service.NewService(config.ClientName, config.ClientScope,
		config.ClientWebsite, customCSS, config.PostFormats, renderer,
		sessionRepo, appRepo, config.SingleInstance)
	handler := service.NewHandler(s, logger, config.StaticDirectory)

	logger.Println("listening on", config.ListenAddress)
	err = http.ListenAndServe(config.ListenAddress, handler)
	if err != nil {
		errExit(err)
	}
}
