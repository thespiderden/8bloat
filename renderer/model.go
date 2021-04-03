package renderer

import (
	"bloat/mastodon"
	"bloat/model"
)

type Context struct {
	HideAttachments  bool
	MaskNSFW         bool
	FluorideMode     bool
	ThreadInNewTab   bool
	DarkMode         bool
	CSRFToken        string
	UserID           string
	AntiDopamineMode bool
	Referrer         string
}

type CommonData struct {
	Title           string
	CustomCSS       string
	CSRFToken       string
	Count           int
	RefreshInterval int
	Target          string
}

type NavData struct {
	CommonData  *CommonData
	User        *mastodon.Account
	PostContext model.PostContext
}

type ErrorData struct {
	*CommonData
	Err        string
	Retry      bool
	SessionErr bool
}

type HomePageData struct {
	*CommonData
}

type SigninData struct {
	*CommonData
}

type RootData struct {
	Title string
}

type TimelineData struct {
	*CommonData
	Title    string
	Type     string
	Instance string
	Statuses []*mastodon.Status
	NextLink string
	PrevLink string
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
	UnreadCount   int
	ReadID        string
	NextLink      string
}

type UserData struct {
	*CommonData
	User      *mastodon.Account
	IsCurrent bool
	Type      string
	Users     []*mastodon.Account
	Statuses  []*mastodon.Status
	NextLink  string
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
	Settings    *model.Settings
	PostFormats []model.PostFormat
}

type FiltersData struct {
	*CommonData
	Filters []*mastodon.Filter
}
