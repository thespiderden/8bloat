package model

import (
	"strings"

	ua "github.com/mileusna/useragent"
)

type PostFormat struct {
	Name string
	Type string
}

type PostContext struct {
	DefaultVisibility string
	DefaultFormat     string
	ReplyContext      *ReplyContext
	Formats           []PostFormat
	UserAgent         ua.UserAgent
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
	if !strings.HasPrefix(r.ReplySubjectHeader, "re: ") {
		return "re: " + r.ReplySubjectHeader
	}

	return r.ReplySubjectHeader
}
