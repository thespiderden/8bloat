package model

import "mastodon"

type Client struct {
	*mastodon.Client
	Session Session
}
