//go:build linux || freebsd || plan9 || openbsd || netbsd || solaris || macos

package conf

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGHUP)

	go func() {
		for {
			<-sigch
			if file == "-" {
				log.Println("recieved sighup, but config is from stdin and cannot be reloaded")
				continue
			}

			f, err := os.Open(file)
			if err != nil {
				log.Println("recieved sighup, error while opening file:", err)
				continue
			}

			lock.RLock()

			err = readConf(f)
			f.Close()
			if err != nil {
				lock.RUnlock()
				log.Println("recieved sighup, error while parsing and applying config:", err)
				continue
			}

			lock.RUnlock()
			log.Println("recieved sighup, reloaded config")
		}
	}()
}
