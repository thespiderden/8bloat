package main

import (
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"web/config"
	"web/renderer"
	"web/repository"
	"web/service"

	_ "github.com/mattn/go-sqlite3"
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

	db, err := sql.Open("sqlite3", config.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sessionRepo, err := repository.NewSessionRepository(db)
	if err != nil {
		log.Fatal(err)
	}

	appRepo, err := repository.NewAppRepository(db)
	if err != nil {
		log.Fatal(err)
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

	s := service.NewService(config.ClientName, config.ClientScope, config.ClientWebsite, renderer, sessionRepo, appRepo)
	s = service.NewAuthService(sessionRepo, appRepo, s)
	s = service.NewLoggingService(logger, s)
	handler := service.NewHandler(s, config.StaticDirectory)

	log.Println("listening on", config.ListenAddress)
	err = http.ListenAndServe(config.ListenAddress, handler)
	if err != nil {
		log.Fatal(err)
	}
}
