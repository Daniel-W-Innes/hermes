package controllers

import (
	"fmt"
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/Daniel-W-Innes/hermes/models"
	"gorm.io/gorm"
	"log"
)

func Login(db *gorm.DB, userLogin *models.UserLogin) (*models.JWT, hermesErrors.HermesError) {
	var user models.User

	db.Where("username = ?", userLogin.Username).First(&user)

	if err := user.CheckPassword([]byte(userLogin.Password)); err != nil {
		return nil, hermesErrors.BadLogin()
	}

	jwt, err := user.GenerateJWT()
	if err != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to generate jwt: %s\n", err))
	}

	return jwt, nil
}

func AddUser(db *gorm.DB, userLogin *models.UserLogin) (*models.JWT, hermesErrors.HermesError) {
	var user models.User

	result := db.Where("username = ?", userLogin.Username).Limit(1).Find(&user)
	if result.Error != nil {
		log.Printf("failed to get user: %s\n", result.Error)
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to get user: %s\n", result.Error))
	}
	if result.RowsAffected > 0 {
		return nil, hermesErrors.UserExists()
	} else {
		user = models.User{
			Username: userLogin.Username,
		}
		err := user.SetPassword([]byte(userLogin.Password))
		if err != nil {
			log.Printf("failed to hash pass: %s\n", err)
			return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to hash pass: %s\n", err))
		}
		db.Create(&user)

		jwt, err := user.GenerateJWT()
		if err != nil {
			return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to generate jwt: %s\n", err))
		}
		return jwt, nil
	}
}
