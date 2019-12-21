package model

import (
	"errors"
)

var (
	ErrAppNotFound = errors.New("app not found")
)

type App struct {
	InstanceDomain string `json:"instance_domain"`
	InstanceURL    string `json:"instance_url"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
}

type AppRepository interface {
	Add(app App) (err error)
	Get(instanceDomain string) (app App, err error)
}
