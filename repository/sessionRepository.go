package repository

import (
	"database/sql"

	"web/model"
)

type sessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) (*sessionRepository, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS session 
		(id varchar, instance_url varchar, access_token varchar)`,
	)
	if err != nil {
		return nil, err
	}

	return &sessionRepository{
		db: db,
	}, nil
}

func (repo *sessionRepository) Add(s model.Session) (err error) {
	_, err = repo.db.Exec("INSERT INTO session VALUES (?, ?, ?)", s.ID, s.InstanceURL, s.AccessToken)
	return
}

func (repo *sessionRepository) Update(sessionID string, accessToken string) (err error) {
	_, err = repo.db.Exec("UPDATE session SET access_token = ? where id = ?", accessToken, sessionID)
	return
}

func (repo *sessionRepository) Get(id string) (s model.Session, err error) {
	rows, err := repo.db.Query("SELECT * FROM session WHERE id = ?", id)
	if err != nil {
		return
	}
	defer rows.Close()

	if !rows.Next() {
		err = model.ErrSessionNotFound
		return
	}

	err = rows.Scan(&s.ID, &s.InstanceURL, &s.AccessToken)
	if err != nil {
		return
	}

	return
}
