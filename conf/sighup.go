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

			log.Println("recieved sighup, reloaded config")
		}
	}()
}
