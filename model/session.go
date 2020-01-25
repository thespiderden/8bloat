package model

import (
	"errors"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Session struct {
	ID             string   `json:"id"`
	InstanceDomain string   `json:"instance_domain"`
	AccessToken    string   `json:"access_token"`
	CSRFToken      string   `json:"csrf_token"`
	Settings       Settings `json:"settings"`
}

type SessionRepository interface {
	Add(session Session) (err error)
	Get(sessionID string) (session Session, err error)
}

func (s Session) IsLoggedIn() bool {
	return len(s.AccessToken) > 0
}
