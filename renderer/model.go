package renderer

import (
	"mastodon"
	"web/model"
)

type NavbarData struct {
	User              *mastodon.Account
	NotificationCount int
}

type TimelineData struct {
	Title       string
	Statuses    []*mastodon.Status
	HasNext     bool
	NextLink    string
	HasPrev     bool
	PrevLink    string
	PostContext model.PostContext
	NavbarData  *NavbarData
}

type ThreadData struct {
	Statuses    []*mastodon.Status
	PostContext model.PostContext
	ReplyMap    map[string][]mastodon.ReplyInfo
	NavbarData  *NavbarData
}

type NotificationData struct {
	Notifications []*mastodon.Notification
	HasNext       bool
	NextLink      string
	NavbarData    *NavbarData
}

type UserData struct {
	User       *mastodon.Account
	Statuses   []*mastodon.Status
	HasNext    bool
	NextLink   string
	NavbarData *NavbarData
}

type AboutData struct {
	NavbarData *NavbarData
}

type EmojiData struct {
	Emojis     []*mastodon.Emoji
	NavbarData *NavbarData
}
