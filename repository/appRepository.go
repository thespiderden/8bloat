package repository

import (
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
	err = repo.db.Set(a.InstanceDomain, a.Marshal())
	return
}

func (repo *appRepository) Get(instanceDomain string) (a model.App, err error) {
	data, err := repo.db.Get(instanceDomain)
	if err != nil {
		err = model.ErrAppNotFound
		return
	}

	err = a.Unmarshal(instanceDomain, data)

	return
}
