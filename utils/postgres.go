package utils

import (
	"fmt"
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/Daniel-W-Innes/hermes/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connection() (*gorm.DB, hermesErrors.HermesError) {

	config, err := models.GetConfig()
	if err != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to get config %s\n", err))
	}

	db, err := gorm.Open(postgres.Open(config.DBConfig.GetPsqlConn()), &gorm.Config{PrepareStmt: true})
	if err != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to initialize db session: %s\n", err))
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to get generic database interface: %s\n", err))
	}

	sqlDB.SetMaxOpenConns(config.DBConfig.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.DBConfig.MaxIdleConns)

	return db, nil
}
