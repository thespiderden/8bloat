package repository

import (
	"web/kv"
	"web/model"
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
	err = repo.db.Set(s.ID, s.Marshal())
	return
}

func (repo *sessionRepository) Update(id string, accessToken string) (err error) {
	data, err := repo.db.Get(id)
	if err != nil {
		return
	}

	var s model.Session
	err = s.Unmarshal(id, data)
	if err != nil {
		return
	}

	s.AccessToken = accessToken

	return repo.db.Set(id, s.Marshal())
}

func (repo *sessionRepository) Get(id string) (s model.Session, err error) {
	data, err := repo.db.Get(id)
	if err != nil {
		err = model.ErrSessionNotFound
		return
	}

	err = s.Unmarshal(id, data)

	return
}
