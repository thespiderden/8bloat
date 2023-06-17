package renderer

import (
	"spiderden.org/8b/model"

	"spiderden.org/masta"
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
	UserCSS          string
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
	User        *masta.Account
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
	Statuses []*masta.Status
	NextLink string
	PrevLink string
}

type ListsData struct {
	*CommonData
	Lists []*masta.List
}

type ListData struct {
	*CommonData
	List           *masta.List
	Accounts       []*masta.Account
	Q              string
	SearchAccounts []*masta.Account
}

type ThreadData struct {
	*CommonData
	Statuses    []*StatusData
	PostContext model.PostContext
}

type StatusData struct {
	*masta.Status
	No          *int
	InReplyToNo *int
	Replies     []ThreadReplyData
	ShowReplies bool
}

type ThreadReplyData struct {
	No int
	masta.ID
}

type QuickReplyData struct {
	*CommonData
	Ancestor    *masta.Status
	Status      *masta.Status
	PostContext model.PostContext
}

type NotificationData struct {
	*CommonData
	Notifications []*masta.Notification
	UnreadCount   int
	ReadID        string
	NextLink      string
}

type UserData struct {
	*CommonData
	User         *masta.Account
	Relationship *masta.Relationship
	IsCurrent    bool
	Type         string
	Users        []*masta.Account
	Statuses     []*masta.Status
	NextLink     string
}

type UserSearchData struct {
	*CommonData
	User     *masta.Account
	Q        string
	Statuses []*masta.Status
	NextLink string
}

type AboutData struct {
	*CommonData
}

type EmojiData struct {
	*CommonData
	Emojis []*masta.Emoji
}

type LikedByData struct {
	*CommonData
	Users    []*masta.Account
	NextLink string
}

type RetweetedByData struct {
	*CommonData
	Users    []*masta.Account
	NextLink string
}

type ReactionsData struct {
	*CommonData
	Reactions []masta.EmojiReaction
}

type SearchData struct {
	*CommonData
	Q        string
	Type     string
	Users    []*masta.Account
	Statuses []*masta.Status
	NextLink string
}

type SettingsData struct {
	*CommonData
	Settings    *model.Settings
	PostFormats []model.PostFormat
}

type FiltersData struct {
	*CommonData
	Filters []*masta.Filter
}

type MuteData struct {
	*CommonData
	User *masta.Account
}
