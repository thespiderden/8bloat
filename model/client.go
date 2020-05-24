package model

import (
	"io"

	"bloat/mastodon"
)

type ClientCtx struct {
	SessionID string
	CSRFToken string
}

type Client struct {
	*mastodon.Client
	Writer  io.Writer
	Ctx     ClientCtx
	Session Session
}
