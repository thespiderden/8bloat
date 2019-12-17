package model

import (
	"errors"
	"strings"
)

var (
	ErrAppNotFound = errors.New("app not found")
)

type App struct {
	InstanceDomain string
	InstanceURL    string
	ClientID       string
	ClientSecret   string
}

type AppRepository interface {
	Add(app App) (err error)
	Get(instanceDomain string) (app App, err error)
}

func (a *App) Marshal() []byte {
	str := a.InstanceURL + "\n" + a.ClientID + "\n" + a.ClientSecret
	return []byte(str)
}

func (a *App) Unmarshal(instanceDomain string, data []byte) error {
	str := string(data)
	lines := strings.Split(str, "\n")
	if len(lines) != 3 {
		return errors.New("invalid data")
	}
	a.InstanceDomain = instanceDomain
	a.InstanceURL = lines[0]
	a.ClientID = lines[1]
	a.ClientSecret = lines[2]
	return nil
}
