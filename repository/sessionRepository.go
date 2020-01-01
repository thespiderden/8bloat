package repository

import (
	"encoding/json"

	"bloat/kv"
	"bloat/model"
)

type sessionRepository struct {
	db *kv.Database
}

func NewSessionRepository(db *kv.Database) *sessionRepository {
	return &sessionRepository{
		db: db,
	}
}

func (repo *sessionRepository) Add(s model.Session) (err error) {
	data, err := json.Marshal(s)
	if err != nil {
		return
	}
	err = repo.db.Set(s.ID, data)
	return
}

func (repo *sessionRepository) Update(id string, accessToken string) (err error) {
	data, err := repo.db.Get(id)
	if err != nil {
		return
	}

	var s model.Session
	err = json.Unmarshal(data, &s)
	if err != nil {
		return
	}

	s.AccessToken = accessToken

	data, err = json.Marshal(s)
	if err != nil {
		return
	}

	return repo.db.Set(id, data)
}

func (repo *sessionRepository) Get(id string) (s model.Session, err error) {
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
