package renderer

import (
	"bloat/mastodon"
	"bloat/model"
)

type Context struct {
	MaskNSFW       bool
	FluorideMode   bool
	ThreadInNewTab bool
	DarkMode       bool
	CSRFToken      string
	UserID         string
}

type HeaderData struct {
	Title             string
	NotificationCount int
	CustomCSS         string
	CSRFToken         string
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
	NextLink    string
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
	NextLink      string
	DarkMode      bool
}

type UserData struct {
	*CommonData
	User     *mastodon.Account
	Type     string
	Users    []*mastodon.Account
	Statuses []*mastodon.Status
	NextLink string
	DarkMode bool
}

type UserSearchData struct {
	*CommonData
	User     *mastodon.Account
	Q        string
	Statuses []*mastodon.Status
	NextLink string
}

type AboutData struct {
	*CommonData
}

type EmojiData struct {
	*CommonData
	Emojis []*mastodon.Emoji
}

type LikedByData struct {
	*CommonData
	Users    []*mastodon.Account
	NextLink string
}

type RetweetedByData struct {
	*CommonData
	Users    []*mastodon.Account
	NextLink string
}

type SearchData struct {
	*CommonData
	Q        string
	Type     string
	Users    []*mastodon.Account
	Statuses []*mastodon.Status
	NextLink string
}

type SettingsData struct {
	*CommonData
	Settings *model.Settings
}
