package renderer

import (
	"mastodon"
)

type TimelinePageTemplateData struct {
	Statuses []*mastodon.Status
	HasNext  bool
	NextLink string
	HasPrev  bool
	PrevLink string
}

func NewTimelinePageTemplateData(statuses []*mastodon.Status, hasNext bool, nextLink string, hasPrev bool,
	prevLink string) *TimelinePageTemplateData {
	return &TimelinePageTemplateData{
		Statuses: statuses,
		HasNext:  hasNext,
		NextLink: nextLink,
		HasPrev:  hasPrev,
		PrevLink: prevLink,
	}
}

type ThreadPageTemplateData struct {
	Status    *mastodon.Status
	Context   *mastodon.Context
	PostReply bool
	ReplyToID string
}

func NewThreadPageTemplateData(status *mastodon.Status, context *mastodon.Context, postReply bool, replyToID string) *ThreadPageTemplateData {
	return &ThreadPageTemplateData{
		Status:    status,
		Context:   context,
		PostReply: postReply,
		ReplyToID: replyToID,
	}
}
