package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"spiderden.org/8b/internal/conf"
	"sync"
)

func init() {
	snowflake.Epoch = conf.SnowflakeEpoch
}

var (
	errInvalidArgument  = errors.New("invalid argument")
	errInvalidSession   = errors.New("invalid session")
	errInvalidCSRFToken = errors.New("invalid csrf token")
)

type Service struct {
	cfg        conf.Configuration
	confch     chan conf.Configuration
	confchonce sync.Once
	servelock  sync.Mutex
	client     *http.Client
	sfnode     *snowflake.Node
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, params, redir := router.Lookup(r.Method, r.URL.Path)
	if redir {
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Header().Add("Location", httprouter.CleanPath(r.URL.Path))
		return
	}
	if h == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	r = r.WithContext(context.WithValue(r.Context(), "conf", s.cfg))
	r = r.WithContext(context.WithValue(r.Context(), "client", s.client))
	r = r.WithContext(context.WithValue(r.Context(), "sfnode", s.sfnode))

	h(w, r, params)
}

func (s *Service) ReplaceConfig(config conf.Configuration) {
	s.confchonce.Do(func() { s.confch = make(chan conf.Configuration) })
	s.confch <- config
}

func (s *Service) StartAndListen(ctx context.Context, config conf.Configuration) error {
	if !s.servelock.TryLock() {
		return errors.New("service already running")
	}

	defer s.servelock.Unlock()

	server := &http.Server{
		Addr:    config.ListenAddress,
		Handler: s,
	}

	s.cfg = config

	var err error
	s.sfnode, err = snowflake.NewNode(int64(config.Node))
	if err != nil {
		return errors.New("unable to create snowflake node: " + err.Error())
	}

	if config.AssetStamp == "random" || config.AssetStamp == "snowflake" {
		s.cfg.AssetStamp = s.sfnode.Generate().Base64()
	}

	errch := make(chan error)
	s.confchonce.Do(func() { s.confch = make(chan conf.Configuration) })
	s.client = newClient(config)

	go func() { errch <- server.ListenAndServe() }()

	for {
		select {
		case err := <-errch:
			return err
		case <-ctx.Done():
			go func() { fmt.Println(server.Shutdown(context.TODO())) }()
			<-errch
			return nil
		case config := <-s.confch:
			go func() { server.Shutdown(context.TODO()) }()
			<-errch

			if config.AssetStamp == "random" || config.AssetStamp == "snowflake" {
				config.AssetStamp = s.sfnode.Generate().Base64()
			}

			s.cfg = config
			s.client = newClient(config)
			s.sfnode, err = snowflake.NewNode(config.Node)
			if err != nil {
				return errors.New("error while reloading config and creating snowflake node: " + err.Error())
			}

			server = &http.Server{
				Addr:    config.ListenAddress,
				Handler: s,
			}

			go func() { errch <- server.ListenAndServe() }()
		}
	}
}
