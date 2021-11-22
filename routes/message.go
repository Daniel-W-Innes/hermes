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

//preHandlerMessage standard handler setup get message from body message par is not nil
func preHandlerMessage(c *fiber.Ctx, message *models.Message) (*gorm.DB, uint, hermesErrors.HermesError) {
	config, err := models.GetConfig()
	if err != nil {
		return nil, 0, hermesErrors.InternalServerError(fmt.Sprintf("failed to get config %s\n", err))
	}

	// check user authorization from authorization header
	authorization := c.Get(fiber.HeaderAuthorization)
	userId, hermesError := utils.ValidateAuth(&config.JWTConfig, authorization)
	if hermesError != nil {
		return nil, 0, hermesError.Wrap("failed on pre handler for message\n")
	}

	db, hermesError := utils.Connection(&config.DBConfig)
	if hermesError != nil {
		return nil, 0, hermesError.Wrap("failed on pre handler for message\n")
	}

	// get user input from body if a destination is provided for it
	if message != nil {
		if err := c.BodyParser(message); err != nil {
			return nil, 0, hermesErrors.UnprocessableEntity(fmt.Sprintf("failed to parser user input: %s\n", err)).Wrap("failed on pre handler for message\n")
		}

		// validate user inputted message
		err := utils.Validate(message)
		if err != nil {
			return nil, 0, err.Wrap("failed on pre handler for message\n")
		}

	}
	return db, userId, nil
}

func addMessage(c *fiber.Ctx) error {
	message := new(models.Message)
	db, userId, hermesError := preHandlerMessage(c, message)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}

	output, hermesError := controllers.AddMessage(db, message, userId)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}
	return c.JSON(output)
}

func deleteMessage(c *fiber.Ctx) error {
	messageId, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	db, userId, hermesError := preHandlerMessage(c, nil)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}

	message, hermesError := controllers.DeleteMessage(db, messageId, userId)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}
	return c.JSON(message)
}

func getMessages(c *fiber.Ctx) error {
	db, userId, hermesError := preHandlerMessage(c, nil)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}

	message, hermesError := controllers.GetMessages(db, userId)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}
	return c.JSON(message)
}

func getMessage(c *fiber.Ctx) error {
	messageId, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	db, userId, hermesError := preHandlerMessage(c, nil)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}

	message, hermesError := controllers.GetMessage(db, messageId, userId)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}
	return c.JSON(message)
}

func editMessage(c *fiber.Ctx) error {
	messageId, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	db, userId, hermesError := preHandlerMessage(c, nil)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}

	message, hermesError := controllers.EditMessage(db, c.BodyParser, messageId, userId)
	if hermesError != nil {
		hermesError.LogPrivate()
		return hermesError
	}
	return c.JSON(message)
}

func Message(app *fiber.App) {
	route := app.Group("/message")

	route.Post("", addMessage)
	route.Delete("/:id", deleteMessage)
	route.Get("", getMessages)
	route.Get("/:id", getMessage)
	route.Post("/:id", editMessage)
}
