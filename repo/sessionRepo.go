package repo

import (
	"encoding/json"

	"bloat/util"
	"bloat/model"
)

type sessionRepo struct {
	db *util.Database
}

func NewSessionRepo(db *util.Database) *sessionRepo {
	return &sessionRepo{
		db: db,
	}
}

func (repo *sessionRepo) Add(s model.Session) (err error) {
	data, err := json.Marshal(s)
	if err != nil {
		return
	}
	err = repo.db.Set(s.ID, data)
	return
}

func (repo *sessionRepo) Get(id string) (s model.Session, err error) {
	data, err := repo.db.Get(id)
	if err != nil {
		err = model.ErrSessionNotFound
		return
	}

	err = json.Unmarshal(data, &s)
	if err != nil {
		return
	}

	return
}

func (repo *sessionRepo) Remove(id string) {
	repo.db.Remove(id)
	return
}
