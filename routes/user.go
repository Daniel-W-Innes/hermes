package routes

import (
	"database/sql"
	"errors"
	"github.com/Daniel-W-Innes/messenger/models"
	"github.com/Daniel-W-Innes/messenger/utils"
	"github.com/gofiber/fiber/v2"
	"log"
)

var BadLogin = fiber.NewError(fiber.StatusUnauthorized, "username or password is not right")
var UserExists = fiber.NewError(fiber.StatusBadRequest, "user already exists")

func login(c *fiber.Ctx) error {
	userLogin := new(models.UserLogin)

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

	user := models.User{Username: userLogin.Username}
	err = user.Get(db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return BadLogin
		}
		log.Printf("failed to get user: %s\n", err)
		return fiber.ErrInternalServerError
	}

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

	user := models.User{Username: userLogin.Username}
	err = user.Get(db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			user = models.User{
				Username: userLogin.Username,
			}
			err = user.SetPassword([]byte(userLogin.Password))
			if err != nil {
				log.Printf("failed to hash pass: %s\n", err)
				return fiber.ErrInternalServerError
			}
			err = user.Insert(db)
			if err != nil {
				log.Printf("failed to insert user: %s\n", err)
				return fiber.ErrInternalServerError
			}

			jwt, err := user.GenerateJWT()
			if err != nil {
				log.Printf("failed to generate jwt: %s\n", err)
				return fiber.ErrInternalServerError
			}
			return c.JSON(jwt)
		}
		log.Printf("failed to get user: %s\n", err)
		return fiber.ErrInternalServerError
	}
	return UserExists
}

func User(app *fiber.App) {
	route := app.Group("/user")

	route.Post("/login", login)
	route.Post("", addUser)

}
