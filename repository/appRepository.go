package repository

import (
	"database/sql"

	"web/model"
)

type appRepository struct {
	db *sql.DB
}

func NewAppRepository(db *sql.DB) (*appRepository, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS app 
		(instance_url varchar, client_id varchar, client_secret varchar)`,
	)
	if err != nil {
		return nil, err
	}

	return &appRepository{
		db: db,
	}, nil
}

func (repo *appRepository) Add(a model.App) (err error) {
	_, err = repo.db.Exec("INSERT INTO app VALUES (?, ?, ?)", a.InstanceURL, a.ClientID, a.ClientSecret)
	return
}

func (repo *appRepository) Update(instanceURL string, clientID string, clientSecret string) (err error) {
	_, err = repo.db.Exec("UPDATE app SET client_id = ?, client_secret = ? where instance_url = ?", clientID, clientSecret, instanceURL)
	return
}

func (repo *appRepository) Get(instanceURL string) (a model.App, err error) {
	rows, err := repo.db.Query("SELECT * FROM app WHERE instance_url = ?", instanceURL)
	if err != nil {
		return
	}
	defer rows.Close()

	if !rows.Next() {
		err = model.ErrAppNotFound
		return
	}

	err = rows.Scan(&a.InstanceURL, &a.ClientID, &a.ClientSecret)
	if err != nil {
		return
	}

	return
}
