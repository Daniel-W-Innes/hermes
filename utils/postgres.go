package utils

import (
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connection() (*sqlx.DB, error) {

	config, err := models.GetConfig()
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("postgres", config.DBConfig.GetPsqlConn())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.DBConfig.MaxOpenConns)
	db.SetMaxIdleConns(config.DBConfig.MaxIdleConns)

	return db, nil
}
