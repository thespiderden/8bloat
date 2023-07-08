package model

import (
	"strings"

	ua "github.com/mileusna/useragent"
	"spiderden.org/masta"
)

type PostFormat struct {
	Name string
	Type string
}

type PostContext struct {
	DefaultVisibility string
	DefaultFormat     string
	ReplyContext      *ReplyContext
	EditContext       *EditContext
	Formats           []PostFormat
	UserAgent         ua.UserAgent
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
