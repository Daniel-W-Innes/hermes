package utils

import (
	"github.com/Daniel-W-Innes/hermes/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connection() (*gorm.DB, error) {

	config, err := models.GetConfig()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(config.DBConfig.GetPsqlConn()), &gorm.Config{PrepareStmt: true})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(config.DBConfig.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.DBConfig.MaxIdleConns)

	return db, nil
}
