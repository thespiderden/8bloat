package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"web/config"
	"web/kv"
	"web/renderer"
	"web/repository"
	"web/service"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	config, err := config.ParseFile("default.conf")
	if err != nil {
		log.Fatal(err)
	}

	if !config.IsValid() {
		log.Fatal("invalid config")
	}

	renderer, err := renderer.NewRenderer(config.TemplatesGlobPattern)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir(config.DatabasePath, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	sessionDB, err := kv.NewDatabse(filepath.Join(config.DatabasePath, "session"))
	if err != nil {
		log.Fatal(err)
	}

	appDB, err := kv.NewDatabse(filepath.Join(config.DatabasePath, "app"))
	if err != nil {
		log.Fatal(err)
	}

	sessionRepo := repository.NewSessionRepository(sessionDB)
	appRepo := repository.NewAppRepository(appDB)

	customCSS := config.CustomCSS
	if !strings.HasPrefix(customCSS, "http://") &&
		!strings.HasPrefix(customCSS, "https://") {
		customCSS = "/static/" + customCSS
	}

	var logger *log.Logger
	if len(config.Logfile) < 1 {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		lf, err := os.Open(config.Logfile)
		if err != nil {
			log.Fatal(err)
		}
		defer lf.Close()
		logger = log.New(lf, "", log.LstdFlags)
	}

	s := service.NewService(config.ClientName, config.ClientScope, config.ClientWebsite,
		customCSS, config.PostFormats, renderer, sessionRepo, appRepo)
	s = service.NewAuthService(sessionRepo, appRepo, s)
	s = service.NewLoggingService(logger, s)
	handler := service.NewHandler(s, config.StaticDirectory)

	log.Println("listening on", config.ListenAddress)
	err = http.ListenAndServe(config.ListenAddress, handler)
	if err != nil {
		log.Fatal(err)
	}
}
