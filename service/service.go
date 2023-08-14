package service

import (
	"context"
	"errors"
	"net/http"

	"spiderden.org/8b/conf"
)

var (
	errInvalidArgument  = errors.New("invalid argument")
	errInvalidSession   = errors.New("invalid session")
	errInvalidCSRFToken = errors.New("invalid csrf token")
)

func StartAndListen(ctx context.Context) error {
	server := &http.Server{
		Addr:    conf.ListenAddress,
		Handler: router,
	}

	errch := make(chan error)
	go func() { errch <- server.ListenAndServe() }()

	select {
	case err := <-errch:
		return err
	case <-ctx.Done():
		server.Shutdown(context.Background())
		return nil
	}
}

func singleInstance() (instance string, ok bool) {
	if len(conf.SingleInstance) > 0 {
		instance = conf.SingleInstance
		ok = true
	}
	return
}
