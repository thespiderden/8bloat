package renderer

import (
	"mastodon"
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
	Status       *mastodon.Status
	Context      *mastodon.Context
	PostReply    bool
	ReplyToID    string
	ReplyContent string
	NavbarData   *NavbarTemplateData
}

func NewThreadPageTemplateData(status *mastodon.Status, context *mastodon.Context, postReply bool, replyToID string, replyContent string, navbarData *NavbarTemplateData) *ThreadPageTemplateData {
	return &ThreadPageTemplateData{
		Status:       status,
		Context:      context,
		PostReply:    postReply,
		ReplyToID:    replyToID,
		ReplyContent: replyContent,
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
