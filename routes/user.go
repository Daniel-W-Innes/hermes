package routes

import (
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/Daniel-W-Innes/hermes/utils"
	"github.com/gofiber/fiber/v2"
	"log"
)

var BadLogin = fiber.NewError(fiber.StatusUnauthorized, "username or password is not right")
var UserExists = fiber.NewError(fiber.StatusBadRequest, "user already exists")

func login(c *fiber.Ctx) error {
	userLogin := new(models.UserLogin)
	user := new(models.User)

	if err := c.BodyParser(userLogin); err != nil {
		return err
	}

	err := utils.Validate(userLogin)
	if err != nil {
		return err
	}

	db, err := utils.Connection()
	if err != nil {
		log.Printf("failed to connect to db: %s\n", err)
		return fiber.ErrInternalServerError
	}

	db.Where("username = ?", userLogin.Username).First(user)

	if err = user.CheckPassword([]byte(userLogin.Password)); err != nil {
		return BadLogin
	}

	jwt, err := user.GenerateJWT()
	if err != nil {
		log.Printf("failed to generate jwt: %s\n", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(jwt)
}

func addUser(c *fiber.Ctx) error {
	userLogin := new(models.UserLogin)
	user := new(models.User)

	if err := c.BodyParser(userLogin); err != nil {
		return err
	}

	err := utils.Validate(userLogin)
	if err != nil {
		return err
	}

	db, err := utils.Connection()
	if err != nil {
		log.Printf("failed to connect to db: %s\n", err)
		return fiber.ErrInternalServerError
	}

	result := db.Where("username = ?", userLogin.Username).Limit(1).Find(user)
	if result.Error != nil {
		log.Printf("failed to get user: %s\n", err)
		return fiber.ErrInternalServerError
	}
	if result.RowsAffected > 0 {
		return UserExists
	} else {
		user = &models.User{
			Username: userLogin.Username,
		}
		err = user.SetPassword([]byte(userLogin.Password))
		if err != nil {
			log.Printf("failed to hash pass: %s\n", err)
			return fiber.ErrInternalServerError
		}
		db.Create(user)

		jwt, err := user.GenerateJWT()
		if err != nil {
			log.Printf("failed to generate jwt: %s\n", err)
			return fiber.ErrInternalServerError
		}
		return c.JSON(jwt)
	}
}

func User(app *fiber.App) {
	route := app.Group("/user")

	route.Post("/login", login)
	route.Post("", addUser)

}
