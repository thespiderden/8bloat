package model

import (
	"errors"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Session struct {
	ID             string   `json:"id"`
	UserID         string   `json:"user_id"`
	InstanceDomain string   `json:"instance_domain"`
	AccessToken    string   `json:"access_token"`
	CSRFToken      string   `json:"csrf_token"`
	Settings       Settings `json:"settings"`
}

type SessionRepo interface {
	Add(session Session) (err error)
	Get(sessionID string) (session Session, err error)
	Remove(sessionID string)
}

func (s Session) IsLoggedIn() bool {
	return len(s.AccessToken) > 0
}
