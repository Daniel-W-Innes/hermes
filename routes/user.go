package routes

import (
	"github.com/Daniel-W-Innes/hermes/controllers"
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/Daniel-W-Innes/hermes/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
)

func preHandlerUser(c *fiber.Ctx, userLogin *models.UserLogin) (*gorm.DB, error) {
	db, err := utils.Connection()
	if err != nil {
		log.Printf("failed to connect to db: %s\n", err)
		return nil, fiber.ErrInternalServerError
	}
	if err := c.BodyParser(userLogin); err != nil {
		return nil, err
	}

	err = utils.Validate(userLogin)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func login(c *fiber.Ctx) error {
	userLogin := new(models.UserLogin)
	db, err := preHandlerUser(c, userLogin)
	if err != nil {
		return err
	}

	message, err := controllers.Login(db, userLogin)
	if err != nil {
		return err
	}
	return c.JSON(message)
}

func addUser(c *fiber.Ctx) error {
	userLogin := new(models.UserLogin)
	db, err := preHandlerUser(c, userLogin)
	if err != nil {
		return err
	}

	message, err := controllers.AddUser(db, userLogin)
	if err != nil {
		return err
	}
	return c.JSON(message)
}

func User(app *fiber.App) {
	route := app.Group("/user")

	route.Post("/login", login)
	route.Post("", addUser)

}
