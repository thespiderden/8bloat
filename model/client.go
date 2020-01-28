package model

import (
	"io"

	"mastodon"
)

type Client struct {
	*mastodon.Client
	Writer  io.Writer
	Session Session
}
