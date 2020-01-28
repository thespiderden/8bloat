package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bloat/config"
	"bloat/kv"
	"bloat/renderer"
	"bloat/repo"
	"bloat/service"
	"bloat/util"
)

var (
	configFile = "bloat.conf"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	opts, _, err := util.Getopts(os.Args, "f:")
	if err != nil {
		log.Fatal(err)
	}

	for _, opt := range opts {
		switch opt.Option {
		case 'f':
			configFile = opt.Value
		}
	}

	config, err := config.ParseFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	if !config.IsValid() {
		log.Fatal("invalid config")
	}

	templatesGlobPattern := filepath.Join(config.TemplatesPath, "*")
	renderer, err := renderer.NewRenderer(templatesGlobPattern)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir(config.DatabasePath, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	sessionDBPath := filepath.Join(config.DatabasePath, "session")
	sessionDB, err := kv.NewDatabse(sessionDBPath)
	if err != nil {
		log.Fatal(err)
	}

	appDBPath := filepath.Join(config.DatabasePath, "app")
	appDB, err := kv.NewDatabse(appDBPath)
	if err != nil {
		log.Fatal(err)
	}

	sessionRepo := repo.NewSessionRepo(sessionDB)
	appRepo := repo.NewAppRepo(appDB)

	customCSS := config.CustomCSS
	if !strings.HasPrefix(customCSS, "http://") &&
		!strings.HasPrefix(customCSS, "https://") {
		customCSS = "/static/" + customCSS
	}

	var logger *log.Logger
	if len(config.LogFile) < 1 {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		lf, err := os.Open(config.LogFile)
		if err != nil {
			log.Fatal(err)
		}
		defer lf.Close()
		logger = log.New(lf, "", log.LstdFlags)
	}

	s := service.NewService(config.ClientName, config.ClientScope,
		config.ClientWebsite, customCSS, config.PostFormats, renderer,
		sessionRepo, appRepo)
	s = service.NewAuthService(sessionRepo, appRepo, s)
	s = service.NewLoggingService(logger, s)
	handler := service.NewHandler(s, config.StaticDirectory)

	logger.Println("listening on", config.ListenAddress)
	err = http.ListenAndServe(config.ListenAddress, handler)
	if err != nil {
		log.Fatal(err)
	}
}
