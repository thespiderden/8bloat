package renderer

import (
	"mastodon"
	"web/model"
)

type NavbarTemplateData struct {
	User              *mastodon.Account
	NotificationCount int
}

func NewNavbarTemplateData(notificationCount int, user *mastodon.Account) *NavbarTemplateData {
	return &NavbarTemplateData{
		NotificationCount: notificationCount,
		User:              user,
	}
}

type TimelinePageTemplateData struct {
	Statuses    []*mastodon.Status
	HasNext     bool
	NextLink    string
	HasPrev     bool
	PrevLink    string
	PostContext model.PostContext
	NavbarData  *NavbarTemplateData
}

func NewTimelinePageTemplateData(statuses []*mastodon.Status, hasNext bool, nextLink string, hasPrev bool,
	prevLink string, postContext model.PostContext, navbarData *NavbarTemplateData) *TimelinePageTemplateData {
	return &TimelinePageTemplateData{
		Statuses:    statuses,
		HasNext:     hasNext,
		NextLink:    nextLink,
		HasPrev:     hasPrev,
		PrevLink:    prevLink,
		PostContext: postContext,
		NavbarData:  navbarData,
	}
}

type ThreadPageTemplateData struct {
	Statuses    []*mastodon.Status
	PostContext model.PostContext
	ReplyMap    map[string][]mastodon.ReplyInfo
	NavbarData  *NavbarTemplateData
}

func NewThreadPageTemplateData(statuses []*mastodon.Status, postContext model.PostContext, replyMap map[string][]mastodon.ReplyInfo, navbarData *NavbarTemplateData) *ThreadPageTemplateData {
	return &ThreadPageTemplateData{
		Statuses:    statuses,
		PostContext: postContext,
		ReplyMap:    replyMap,
		NavbarData:  navbarData,
	}
}

type NotificationPageTemplateData struct {
	Notifications []*mastodon.Notification
	HasNext       bool
	NextLink      string
	NavbarData    *NavbarTemplateData
}

func NewNotificationPageTemplateData(notifications []*mastodon.Notification, hasNext bool, nextLink string, navbarData *NavbarTemplateData) *NotificationPageTemplateData {
	return &NotificationPageTemplateData{
		Notifications: notifications,
		HasNext:       hasNext,
		NextLink:      nextLink,
		NavbarData:    navbarData,
	}
}

type UserPageTemplateData struct {
	User       *mastodon.Account
	Statuses   []*mastodon.Status
	HasNext    bool
	NextLink   string
	NavbarData *NavbarTemplateData
}

func NewUserPageTemplateData(user *mastodon.Account, statuses []*mastodon.Status, hasNext bool, nextLink string, navbarData *NavbarTemplateData) *UserPageTemplateData {
	return &UserPageTemplateData{
		User:       user,
		Statuses:   statuses,
		HasNext:    hasNext,
		NextLink:   nextLink,
		NavbarData: navbarData,
	}
}

type AboutPageTemplateData struct {
	NavbarData *NavbarTemplateData
}

func NewAboutPageTemplateData(navbarData *NavbarTemplateData) *AboutPageTemplateData {
	return &AboutPageTemplateData{
		NavbarData: navbarData,
	}
}

type EmojiPageTemplateData struct {
	NavbarData *NavbarTemplateData
	Emojis     []*mastodon.Emoji
}

func NewEmojiPageTemplateData(navbarData *NavbarTemplateData, emojis []*mastodon.Emoji) *EmojiPageTemplateData {
	return &EmojiPageTemplateData{
		NavbarData: navbarData,
		Emojis:     emojis,
	}
}
