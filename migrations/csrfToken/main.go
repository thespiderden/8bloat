package main

import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"bloat/config"
	"bloat/kv"
	"bloat/repository"
	"bloat/util"
)

var (
	configFile = "bloat.conf"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func getKeys(sessionRepoPath string) (keys []string, err error) {
	f, err := os.Open(sessionRepoPath)
	if err != nil {
		return
	}
	return f.Readdirnames(0)
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

	sessionRepoPath := filepath.Join(config.DatabasePath, "session")
	sessionDB, err := kv.NewDatabse(sessionRepoPath)
	if err != nil {
		log.Fatal(err)
	}

	sessionRepo := repository.NewSessionRepository(sessionDB)

	sessionIds, err := getKeys(sessionRepoPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, id := range sessionIds {
		s, err := sessionRepo.Get(id)
		if err != nil {
			log.Fatal(err)
		}
		s.CSRFToken, err = util.NewCSRFToken()
		if err != nil {
			log.Fatal(err)
		}
		err = sessionRepo.Add(s)
		if err != nil {
			log.Fatal(err)
		}
	}

}
