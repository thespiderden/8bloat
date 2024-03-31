package main

import (
	"context"
	"errors"
	"flag"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"spiderden.org/8b/internal/conf"
	"spiderden.org/8b/internal/service"
	"syscall"
)

var (
	file      = flag.String("f", "", `config file, use a dash for stdin`)
	writeConf = flag.Bool("wc", false, `write a sample configuration file to stdout`)
)

func main() {
	flag.Parse()

	if *writeConf && (*file != "") {
		log.Fatal("cannot use -f and -wc at the same time")
	}

	if *writeConf {
		_, err := os.Stdout.Write(DefaultConfig)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *file == "" {
		var path string
		for _, path = range []string{"8bloat.conf", "/etc/8bloat.conf", "bloat.conf", "/etc/bloat.conf"} {
			stat, err := os.Stat(path)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				log.Fatal("error searching for config f: ", err)
			}

			if !stat.IsDir() {
				file = &path
				break
			}
		}

		if *file == "" {
			log.Fatal("exhausted default config search, please specify your own")
		}
	}

	var cfg conf.Configuration
	var err error
	if *file == "-" {
		if cfg, err = readConf(os.Stdin); err != nil {
			log.Fatal("configuration error", err)
		}
	} else {
		f, err := os.Open(*file)
		if err != nil {
			log.Fatal(err)
		}

		cfg, err = readConf(f)
		if err != nil {
			log.Fatal("configuration error", err)
		}
	}

	ctx, ctxf := context.WithCancel(context.Background())
	errch := make(chan error)

	log.Println("starting service on", cfg.ListenAddress)
	serv := &service.Service{}

	go func() {
		errch <- serv.StartAndListen(ctx, cfg)
	}()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGHUP)

	for {
		select {
		case sig := <-sigch:
			switch sig {
			case syscall.SIGHUP:
				if *file == "-" {
					log.Println("recieved sighup, but config is from stdin and cannot be reloaded")
					continue
				}

				f, err := os.Open(*file)
				if err != nil {
					log.Println("recieved sighup, error while opening config file:", err)
					continue
				}

				cfg, err := readConf(f)
				f.Close()
				if err != nil {
					log.Println("recieved sighup, error while parsing and applying config:", err)
					continue
				}

				serv.ReplaceConfig(cfg)

				log.Println("recieved sighup, reloaded config")
			case os.Interrupt:
				log.Println("got signal to terminate, gracefully stopping")
				ctxf()
				<-errch
				os.Exit(0)
			}
		case err := <-errch:
			log.Fatal(err)
		}
	}
}
