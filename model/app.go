package model

import "errors"

var (
	ErrAppNotFound = errors.New("app not found")
)

type App struct {
	InstanceURL  string
	ClientID     string
	ClientSecret string
}

type AppRepository interface {
	Add(app App) (err error)
	Update(instanceURL string, clientID string, clientSecret string) (err error)
	Get(instanceURL string) (app App, err error)
}
