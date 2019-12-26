package renderer

import (
	"mastodon"
	"web/model"
)

type HeaderData struct {
	Title             string
	NotificationCount int
	CustomCSS         string
}

type NavbarData struct {
	User              *mastodon.Account
	NotificationCount int
}

type CommonData struct {
	HeaderData *HeaderData
	NavbarData *NavbarData
}

type ErrorData struct {
	*CommonData
	Error string
}

type HomePageData struct {
	*CommonData
}

type SigninData struct {
	*CommonData
}

type TimelineData struct {
	*CommonData
	Title       string
	Statuses    []*mastodon.Status
	HasNext     bool
	NextLink    string
	HasPrev     bool
	PrevLink    string
	PostContext model.PostContext
}

type ThreadData struct {
	*CommonData
	Statuses    []*mastodon.Status
	PostContext model.PostContext
	ReplyMap    map[string][]mastodon.ReplyInfo
}

type NotificationData struct {
	*CommonData
	Notifications []*mastodon.Notification
	HasNext       bool
	NextLink      string
}

type UserData struct {
	*CommonData
	User       *mastodon.Account
	Statuses   []*mastodon.Status
	HasNext    bool
	NextLink   string
}

type AboutData struct {
	*CommonData
}

type EmojiData struct {
	*CommonData
	Emojis     []*mastodon.Emoji
}

type LikedByData struct {
	*CommonData
	Users []*mastodon.Account
	HasNext    bool
	NextLink   string
}

type RetweetedByData struct {
	*CommonData
	Users []*mastodon.Account
	HasNext    bool
	NextLink   string
}
