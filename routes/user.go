package routes

import (
	"fmt"
	"github.com/Daniel-W-Innes/hermes/controllers"
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/Daniel-W-Innes/hermes/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func preHandlerUser(c *fiber.Ctx, userLogin *models.UserLogin) (*models.Config, *gorm.DB, hermesErrors.HermesError) {
	config, err := models.GetConfig()
	if err != nil {
		return nil, nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to get config %s\n", err))
	}

	db, err := utils.Connection(&config.DBConfig)
	if err != nil {
		return nil, nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to connect to db: %s\n", err)).Wrap("failed on pre handler for user\n")
	}
	if err := c.BodyParser(userLogin); err != nil {
		return nil, nil, hermesErrors.UnprocessableEntity(fmt.Sprintf("failed to parser user input: %s\n", err)).Wrap("failed on pre handler for user\n")
	}

	hermesError := utils.Validate(userLogin)
	if hermesError != nil {
		return nil, nil, hermesError.Wrap("failed on pre handler for user\n")
	}

	return config, db, nil
}

func login(c *fiber.Ctx) error {
	userLogin := new(models.UserLogin)
	config, db, err := preHandlerUser(c, userLogin)
	if err != nil {
		return err
	}

	message, hermesError := controllers.Login(db, config, userLogin)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}
	return c.JSON(message)
}

func addUser(c *fiber.Ctx) error {
	userLogin := new(models.UserLogin)
	config, db, err := preHandlerUser(c, userLogin)
	if err != nil {
		return err
	}

	message, hermesError := controllers.AddUser(db, config, userLogin)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}
	return c.JSON(message)
}

func User(app *fiber.App) {
	route := app.Group("/user")

	route.Post("/login", login)
	route.Post("", addUser)

}
