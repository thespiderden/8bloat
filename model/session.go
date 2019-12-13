package model

import "errors"

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Session struct {
	ID          string
	InstanceURL string
	AccessToken string
}

type SessionRepository interface {
	Add(session Session) (err error)
	Update(sessionID string, accessToken string) (err error)
	Get(sessionID string) (session Session, err error)
}

func (s Session) IsLoggedIn() bool {
	return len(s.AccessToken) > 0
}
