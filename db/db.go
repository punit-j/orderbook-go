package db

import (
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"orders-manager/config"
	"orders-manager/errors"
	"orders-manager/models"

	logger "github.com/sirupsen/logrus"
)

var (
	db *gorm.DB
)

func Init() (err error) {
	dbURI := config.ReadEnvString("DB_URI")
	db, err = gorm.Open(postgres.Open(dbURI), &gorm.Config{})
	if err != nil {
		logger.WithField("error", err).Fatal("failed to connect database")
		return errors.ErrFailedToConnectDB
	}

	// models to be intialized here
	initModels()

	logger.Info("Database connection established")

	return nil
}

func initModels() {
	logger.Info("Initializing models....")
	models.InitModels(db)
}
