package repository

import (
	"encoding/json"
	"web/kv"
	"web/model"
)

type appRepository struct {
	db *kv.Database
}

func NewAppRepository(db *kv.Database) *appRepository {
	return &appRepository{
		db: db,
	}
}

func (repo *appRepository) Add(a model.App) (err error) {
	data, err := json.Marshal(a)
	if err != nil {
		return
	}
	err = repo.db.Set(a.InstanceDomain, data)
	return
}

func (repo *appRepository) Get(instanceDomain string) (a model.App, err error) {
	data, err := repo.db.Get(instanceDomain)
	if err != nil {
		err = model.ErrAppNotFound
		return
	}

	err = json.Unmarshal(data, &a)
	if err != nil {
		return
	}

	return
}
