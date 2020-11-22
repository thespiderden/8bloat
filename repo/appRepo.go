package repo

import (
	"encoding/json"

	"bloat/util"
	"bloat/model"
)

type appRepo struct {
	db *util.Database
}

func NewAppRepo(db *util.Database) *appRepo {
	return &appRepo{
		db: db,
	}
}

func (repo *appRepo) Add(a model.App) (err error) {
	data, err := json.Marshal(a)
	if err != nil {
		return
	}
	err = repo.db.Set(a.InstanceDomain, data)
	return
}

func (repo *appRepo) Get(instanceDomain string) (a model.App, err error) {
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
