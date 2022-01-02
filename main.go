package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"bloat/config"
	"bloat/renderer"
	"bloat/repo"
	"bloat/service"
	"bloat/util"
)

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

	s := service.NewService(config.ClientName, config.ClientScope,
		config.ClientWebsite, customCSS, config.SingleInstance,
		config.PostFormats, renderer, sessionRepo, appRepo)
	handler := service.NewHandler(s, logger, config.StaticDirectory)

	logger.Println("listening on", config.ListenAddress)
	err = http.ListenAndServe(config.ListenAddress, handler)
	if err != nil {
		errExit(err)
	}
}
