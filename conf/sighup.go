package conf

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	changedChs     []chan Configuration
	changedChsLock sync.RWMutex
)

func Changed() chan Configuration {
	changed := make(chan Configuration)
	changedChsLock.RLock()
	changedChs = append(changedChs, changed)
	changedChsLock.RUnlock()
	return changed
}

func init() {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGHUP)

	go func() {
		for {
			<-sigch
			if *file == "-" {
				log.Println("recieved sighup, but config is from stdin and cannot be reloaded")
				continue
			}

			f, err := os.Open(*file)
			if err != nil {
				log.Println("recieved sighup, error while opening file:", err)
				continue
			}

			err = readConf(f)
			if err != nil {
				log.Println("recieved sighup, error while parsing and applying config:", err)
				continue
			}

			changedChsLock.Lock()
			conf := Get()
			for _, v := range changedChs {
				v <- *conf
			}
			changedChsLock.Unlock()

			log.Println("recieved sighup, reloaded config")
		}
	}()
}
