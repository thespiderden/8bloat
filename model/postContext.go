package model

type PostFormat struct {
	Name string
	Type string
}

type PostContext struct {
	DefaultVisibility string
	ReplyContext      *ReplyContext
	Formats           []PostFormat
}

type ReplyContext struct {
	InReplyToID   string
	InReplyToName string
	ReplyContent  string
}