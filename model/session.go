package model

import (
	"errors"
	"strings"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Session struct {
	ID             string
	InstanceDomain string
	AccessToken    string
}

type SessionRepository interface {
	Add(session Session) (err error)
	Update(sessionID string, accessToken string) (err error)
	Get(sessionID string) (session Session, err error)
}

func (s Session) IsLoggedIn() bool {
	return len(s.AccessToken) > 0
}

func (s *Session) Marshal() []byte {
	str := s.InstanceDomain + "\n" + s.AccessToken
	return []byte(str)
}

func (s *Session) Unmarshal(id string, data []byte) error {
	str := string(data)
	lines := strings.Split(str, "\n")

	size := len(lines)
	if size == 1 {
		s.InstanceDomain = lines[0]
	} else if size == 2 {
		s.InstanceDomain = lines[0]
		s.AccessToken = lines[1]
	} else {
		return errors.New("invalid data")
	}

	s.ID = id
	return nil
}
