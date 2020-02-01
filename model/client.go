package model

import (
	"io"

	"bloat/mastodon"
)

type Client struct {
	*mastodon.Client
	Writer  io.Writer
	Session Session
}
