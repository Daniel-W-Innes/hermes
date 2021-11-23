package controllers

import (
	"fmt"
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/Daniel-W-Innes/hermes/models"
	"gorm.io/gorm"
)

// Login check password and get jwt for an existing user
func Login(db *gorm.DB, config *models.Config, userLogin *models.UserLogin) (*models.JWT, hermesErrors.HermesError) {
	var user models.User

	// get user by username
	db.Where("username = ?", userLogin.Username).First(&user)

	// check password against db row
	if err := user.CheckPassword(&config.PasswordConfig, []byte(userLogin.Password)); err != nil {
		return nil, hermesErrors.BadLogin()
	}

	// generate jwt for user
	jwt, err := user.GenerateJWT(&config.JWTConfig)
	if err != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to generate jwt: %s\n", err))
	}

	return jwt, nil
}

// AddUser add user new user and return a new jwt for the user
func AddUser(db *gorm.DB, config *models.Config, userLogin *models.UserLogin) (*models.JWT, hermesErrors.HermesError) {
	var user models.User

	// check if user exists by username
	result := db.Where("username = ?", userLogin.Username).Limit(1).Find(&user)
	if result.Error != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to get user: %s\n", result.Error))
	}

	// if user does not exist
	if result.RowsAffected > 0 {
		return nil, hermesErrors.UserExists()
	} else {
		// create new user
		user = models.User{
			Username: userLogin.Username,
		}
		// set users password to hashed user password
		err := user.SetPassword(&config.PasswordConfig, []byte(userLogin.Password))
		if err != nil {
			return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to hash pass: %s\n", err))
		}

		// create user in db
		result := db.Create(&user)
		if result.Error != nil {
			return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to create user: %s\n", result.Error))
		}

		// generate jwt for user
		jwt, err := user.GenerateJWT(&config.JWTConfig)
		if err != nil {
			return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to generate jwt: %s\n", err))
		}
		return jwt, nil
	}
}
