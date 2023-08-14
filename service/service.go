package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"net/http"

	"github.com/bwmarrin/snowflake"
	"spiderden.org/8b/conf"
)

func init() {
	snowflake.Epoch = 1665230888000
}

var (
	errInvalidArgument  = errors.New("invalid argument")
	errInvalidSession   = errors.New("invalid session")
	errInvalidCSRFToken = errors.New("invalid csrf token")
)

func StartAndListen(ctx context.Context) error {
	node, err := snowflake.NewNode(int64(conf.SFNodeID))
	if err != nil {
		return err
	}

	// random for backwards compatibility
	if conf.AssetStamp == "random" || conf.AssetStamp == "snowflake" {
		// We do this to get shorter IDs that don't include empty bytes.
		idSlice := &bytes.Buffer{}
		binary.Write(idSlice, binary.LittleEndian, node.Generate())
		conf.AssetStamp = "." + base64.RawURLEncoding.EncodeToString(idSlice.Bytes())
	}

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
