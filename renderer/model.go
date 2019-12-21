package renderer

import (
	"mastodon"
	"web/model"
)

type NavbarTemplateData struct {
	NotificationCount int
}

func NewNavbarTemplateData(notificationCount int) *NavbarTemplateData {
	return &NavbarTemplateData{
		NotificationCount: notificationCount,
	}
}

type TimelinePageTemplateData struct {
	Statuses   []*mastodon.Status
	HasNext    bool
	NextLink   string
	HasPrev    bool
	PrevLink   string
	NavbarData *NavbarTemplateData
}

func NewTimelinePageTemplateData(statuses []*mastodon.Status, hasNext bool, nextLink string, hasPrev bool,
	prevLink string, navbarData *NavbarTemplateData) *TimelinePageTemplateData {
	return &TimelinePageTemplateData{
		Statuses:   statuses,
		HasNext:    hasNext,
		NextLink:   nextLink,
		HasPrev:    hasPrev,
		PrevLink:   prevLink,
		NavbarData: navbarData,
	}
}

type ThreadPageTemplateData struct {
	Statuses     []*mastodon.Status
	ReplyContext *model.ReplyContext
	ReplyMap     map[string][]mastodon.ReplyInfo
	NavbarData   *NavbarTemplateData
}

func NewThreadPageTemplateData(statuses []*mastodon.Status, replyContext *model.ReplyContext, replyMap map[string][]mastodon.ReplyInfo, navbarData *NavbarTemplateData) *ThreadPageTemplateData {
	return &ThreadPageTemplateData{
		Statuses:     statuses,
		ReplyContext: replyContext,
		ReplyMap:     replyMap,
		NavbarData:   navbarData,
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
