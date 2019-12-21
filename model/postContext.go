package model

type PostContext struct {
	DefaultVisibility string
	ReplyContext      *ReplyContext
}

type ReplyContext struct {
	InReplyToID   string
	InReplyToName string
	ReplyContent  string
}
