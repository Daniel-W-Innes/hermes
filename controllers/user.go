package controllers

import (
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
)

var BadLogin = fiber.NewError(fiber.StatusUnauthorized, "username or password is not right")
var UserExists = fiber.NewError(fiber.StatusBadRequest, "user already exists")

func Login(db *gorm.DB, userLogin *models.UserLogin) (interface{}, error) {
	var user models.User

	db.Where("username = ?", userLogin.Username).First(&user)

	if err := user.CheckPassword([]byte(userLogin.Password)); err != nil {
		return nil, BadLogin
	}

	jwt, err := user.GenerateJWT()
	if err != nil {
		log.Printf("failed to generate jwt: %s\n", err)
		return nil, fiber.ErrInternalServerError
	}

	return jwt, nil
}

func AddUser(db *gorm.DB, userLogin *models.UserLogin) (interface{}, error) {
	var user models.User

	result := db.Where("username = ?", userLogin.Username).Limit(1).Find(&user)
	if result.Error != nil {
		log.Printf("failed to get user: %s\n", result.Error)
		return nil, fiber.ErrInternalServerError
	}
	if result.RowsAffected > 0 {
		return nil, UserExists
	} else {
		user = models.User{
			Username: userLogin.Username,
		}
		err := user.SetPassword([]byte(userLogin.Password))
		if err != nil {
			log.Printf("failed to hash pass: %s\n", err)
			return nil, fiber.ErrInternalServerError
		}
		db.Create(&user)

		jwt, err := user.GenerateJWT()
		if err != nil {
			log.Printf("failed to generate jwt: %s\n", err)
			return nil, fiber.ErrInternalServerError
		}
		return jwt, nil
	}
}
