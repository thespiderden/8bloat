package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"spiderden.org/8b/conf"
	"spiderden.org/8b/service"
)

func main() {
	ctx, ctxf := context.WithCancel(context.Background())
	errch := make(chan error)

	log.Println("starting service on", conf.Get().ListenAddress)
	go func() {
		errch <- service.StartAndListen(ctx)
	}()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

	select {
	case err := <-errch:
		log.Fatal(err)
	case <-sigch:
		log.Println("got signal to terminate, gracefully stopping")
		ctxf()
		<-errch
	}
}
