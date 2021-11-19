package routes

import (
	"github.com/Daniel-W-Innes/hermes/controllers"
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/Daniel-W-Innes/hermes/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
)

func preHandlerMessage(c *fiber.Ctx, message *models.Message) (*gorm.DB, uint, error) {
	authorization := c.Get(fiber.HeaderAuthorization)
	userId, err := utils.ValidateAuth(authorization)
	if err != nil {
		return nil, 0, err
	}

	db, err := utils.Connection()
	if err != nil {
		log.Printf("failed to connect to db: %s\n", err)
		return nil, 0, fiber.ErrInternalServerError
	}

	if message != nil {
		if err := c.BodyParser(message); err != nil {
			return nil, 0, err
		}

		err := utils.Validate(message)
		if err != nil {
			return nil, 0, err
		}

	}
	return db, userId, nil
}

func addMessage(c *fiber.Ctx) error {
	message := new(models.Message)
	db, userId, err := preHandlerMessage(c, message)
	if err != nil {
		return err
	}

	return c.JSON(controllers.AddMessage(db, message, userId))
}

func deleteMessage(c *fiber.Ctx) error {
	messageId, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	db, userId, err := preHandlerMessage(c, nil)
	if err != nil {
		return err
	}

	message, err := controllers.DeleteMessage(db, messageId, userId)
	if err != nil {
		return err
	}
	return c.JSON(message)
}

func getMessages(c *fiber.Ctx) error {
	db, userId, err := preHandlerMessage(c, nil)
	if err != nil {
		return err
	}

	message, err := controllers.GetMessages(db, userId)
	if err != nil {
		return err
	}
	return c.JSON(message)
}

func getMessage(c *fiber.Ctx) error {
	messageId, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	db, userId, err := preHandlerMessage(c, nil)
	if err != nil {
		return err
	}

	message, err := controllers.GetMessage(db, messageId, userId)
	if err != nil {
		return err
	}
	return c.JSON(message)
}

func editMessage(c *fiber.Ctx) error {
	messageId, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	db, userId, err := preHandlerMessage(c, nil)
	if err != nil {
		return err
	}

	message, err := controllers.EditMessage(db, c.BodyParser, messageId, userId)
	if err != nil {
		return err
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
