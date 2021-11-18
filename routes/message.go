package routes

import (
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/Daniel-W-Innes/hermes/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm/clause"
	"log"
)

func addMessage(c *fiber.Ctx) error {
	message := new(models.Message)

	if err := c.BodyParser(message); err != nil {
		return err
	}

	err := utils.Validate(message)
	if err != nil {
		return err
	}

	db, err := utils.Connection()
	if err != nil {
		log.Printf("failed to connect to db: %s\n", err)
		return fiber.ErrInternalServerError
	}

	authorization := c.Get(fiber.HeaderAuthorization)
	userId, err := utils.ValidateAuth(authorization)
	if err != nil {
		return err
	}

	message.OwnerID = userId

	message.Check()

	db.Clauses(clause.Returning{}).Save(message)

	return c.JSON(fiber.Map{"id": message.ID})
}

func deleteMessage(c *fiber.Ctx) error {
	messageId, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	db, err := utils.Connection()
	if err != nil {
		log.Printf("failed to connect to db: %s\n", err)
		return fiber.ErrInternalServerError
	}

	authorization := c.Get(fiber.HeaderAuthorization)
	userId, err := utils.ValidateAuth(authorization)
	if err != nil {
		return err
	}

	result := db.Where("id = ?", messageId).Where("owner_id = ?", userId).Delete(&models.Message{})

	if result.Error != nil {
		log.Printf("failed to delete message: %s\n", result.Error)
		return fiber.ErrInternalServerError
	}
	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "message does not exits or token is not the owner")
	}
	return c.SendString("message deleted")
}

func getMessages(c *fiber.Ctx) error {
	var messages []models.Message
	var messagesFromAssociation []models.Message

	authorization := c.Get(fiber.HeaderAuthorization)
	userId, err := utils.ValidateAuth(authorization)
	if err != nil {
		return err
	}

	db, err := utils.Connection()
	if err != nil {
		log.Printf("failed to connect to db: %s\n", err)
		return fiber.ErrInternalServerError
	}

	db.Where("owner_id = ?", userId).Find(&messages)

	err = db.Model(&models.Message{}).Where("id = ?", userId).Association("Recipients").Find(&messagesFromAssociation)

	copy(messages[len(messages):], messagesFromAssociation)

	if err != nil {
		log.Printf("failed to get messages %s\n", err)
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{
		"messages": messages,
	})
}

func getMessage(c *fiber.Ctx) error {
	var message models.Message

	authorization := c.Get(fiber.HeaderAuthorization)
	userId, err := utils.ValidateAuth(authorization)
	if err != nil {
		return err
	}

	messageId, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	db, err := utils.Connection()
	if err != nil {
		log.Printf("failed to connect to db: %s\n", err)
		return fiber.ErrInternalServerError
	}

	result := db.Where("id = ?", messageId).Where("owner_id = ?", userId).Limit(1).Find(&message)

	if result.Error != nil {
		log.Printf("failed to delete message: %s\n", result.Error)
		return fiber.ErrInternalServerError
	}
	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "message not found")
	}
	return c.JSON(message)
}

func editMessage(c *fiber.Ctx) error {
	var message models.Message

	authorization := c.Get(fiber.HeaderAuthorization)
	userId, err := utils.ValidateAuth(authorization)
	if err != nil {
		return err
	}

	messageId, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	db, err := utils.Connection()
	if err != nil {
		log.Printf("failed to connect to db: %s\n", err)
		return fiber.ErrInternalServerError
	}

	result := db.Where("id = ?", messageId).Where("owner_id = ?", userId).Limit(1).Find(&message)
	if result.Error != nil {
		log.Printf("failed to delete message: %s\n", result.Error)
		return fiber.ErrInternalServerError
	}
	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "message not found")
	}

	if err := c.BodyParser(&message); err != nil {
		return err
	}

	message.Check()

	db.Save(&message)

	if err != nil {
		log.Printf("failed to update message %s\n", err)
		return fiber.ErrInternalServerError
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
