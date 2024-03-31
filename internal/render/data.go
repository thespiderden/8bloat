package render

import (
	"io"
	"spiderden.org/8b/internal/conf"
	"strings"

	"spiderden.org/masta"
)

type Context struct {
	Settings   Settings
	CSRFToken  string
	UserID     string
	Referrer   string
	W          io.Writer
	Pagination *masta.Pagination
	Conf       *conf.Configuration

	next            string
	refreshInterval int
	count           int
	target          string
	title           string
}

func (c *Context) RefreshInterval() int {
	return c.refreshInterval
}

func (c *Context) Next() string {
	return c.next
}

func (c *Context) Count() int {
	return c.count
}

func (c *Context) Target() string {
	return c.target
}

func (c *Context) Title() string {
	return c.title
}

type NavData struct {
	Context     *Context
	User        *masta.Account
	PostContext PostContext
}

type ErrorData struct {
	*Context
	Err        string
	Retry      bool
	SessionErr bool
}

type HomePageData struct {
	*Context
}

type SigninData struct {
	*Context
}

type RootData struct {
	*Context
	Title string
}

type TimelineData struct {
	Title    string
	Type     string
	Instance string
	Statuses []*masta.Status
	NextLink string
	PrevLink string
}

type ListsData struct {
	*Context
	Lists []*masta.List
}

type ListData struct {
	List           *masta.List
	Accounts       []*masta.Account
	Q              string
	SearchAccounts []*masta.Account
}

type ThreadData struct {
	Statuses    []*StatusData
	PostContext PostContext
}

type StatusData struct {
	*masta.Status
	No          *int
	InReplyToNo *int
	Replies     []ThreadReplyData
	ShowReplies bool
	History     bool
}

type ThreadReplyData struct {
	No int
	masta.ID
}

type QuickReplyData struct {
	Ancestor    *masta.Status
	Status      *masta.Status
	PostContext PostContext
}

type NotificationData struct {
	Notifications []*masta.Notification
	UnmarkedCount int
	ReadID        string
	NextLink      string
}

type UserData struct {
	User         *masta.Account
	Relationship *masta.Relationship
	IsCurrent    bool
	Type         string
	Users        []*masta.Account
	Statuses     []*masta.Status
	NextLink     string
}

type UserSearchData struct {
	User     *masta.Account
	Q        string
	Statuses []*masta.Status
	NextLink string
}

type EmojiData struct {
	Emojis []*masta.Emoji
}

type LikedByData struct {
	Users    []*masta.Account
	NextLink string
}

type RetweetedByData struct {
	Users    []*masta.Account
	NextLink string
}

type ReactionsData struct {
	Reactions []masta.EmojiReaction
}

type SearchData struct {
	Q        string
	Type     string
	Users    []*masta.Account
	Statuses []*masta.Status
	NextLink string
}

type SettingsData struct {
	Settings    *Settings
	PostFormats []conf.PostFormat
}

type FiltersData struct {
	Filters []*masta.Filter
}

type MuteData struct {
	User *masta.Account
}

type PostContext struct {
	DefaultVisibility string
	DefaultFormat     string
	ReplyContext      *ReplyContext
	EditContext       *EditContext
	Formats           []conf.PostFormat
	Pleroma           bool
}

type ProfileData struct {
	User *masta.Account
}

type EditContext struct {
	Source *masta.Source
	Status *masta.Status
}

type ReplyContext struct {
	InReplyToID        string
	InReplyToName      string
	QuickReply         bool
	ReplyContent       string
	ReplySubjectHeader string
	ForceVisibility    bool
}

func (r *ReplyContext) ReifiedSubjectHeader() string {
	sh := r.ReplySubjectHeader
	if (sh != "") && (!strings.HasPrefix(sh, "re: ")) {
		return "re: " + sh
	}

	return sh
}

type Settings struct {
	DefaultVisibility     string `json:"dv,omitempty"`
	DefaultFormat         string `json:"df,omitempty"`
	CopyScope             bool   `json:"cs,omitempty"`
	ThreadInNewTab        bool   `json:"tnt,omitempty"`
	HideAttachments       bool   `json:"ha,omitempty"`
	MaskNSFW              bool   `json:"mn,omitempty"`
	NotificationInterval  int    `json:"ni,omitempty"`
	FluorideMode          bool   `json:"fm,omitempty"`
	DarkMode              bool   `json:"dm,omitempty"`
	AntiDopamineMode      bool   `json:"adm,omitempty"`
	HideUnsupportedNotifs bool   `json:"hun,omitempty"`
	CSS                   string `json:"css,omitempty"`
	Stamp                 string `json:"stamp,omitempty"`
}

func NewSettings() *Settings {
	return &Settings{
		DefaultVisibility:     "public",
		DefaultFormat:         "",
		CopyScope:             true,
		ThreadInNewTab:        false,
		HideAttachments:       false,
		MaskNSFW:              true,
		NotificationInterval:  0,
		FluorideMode:          false,
		DarkMode:              false,
		AntiDopamineMode:      false,
		HideUnsupportedNotifs: false,
		CSS:                   "",
	}
}
