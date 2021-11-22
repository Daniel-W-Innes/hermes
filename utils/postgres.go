package utils

import (
	"fmt"
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/Daniel-W-Innes/hermes/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connection get db connection from connection pool
func Connection(config *models.DBConfig) (*gorm.DB, hermesErrors.HermesError) {
	// open postgres connection with cashed prepare statements
	db, err := gorm.Open(postgres.Open(config.GetPsqlConn()), &gorm.Config{PrepareStmt: true})
	if err != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to initialize db session: %s\n", err))
	}

	// setup db connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to get generic database interface: %s\n", err))
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)

	return db, nil
}
