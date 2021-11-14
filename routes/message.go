package routes

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/Daniel-W-Innes/hermes/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
	"log"
	"strconv"
)

var MissingMessageID = errors.New("missing message id in prams")

func addMessage(c *fiber.Ctx) error {
	messageBundle := new(models.MessageBundle)

	if err := c.BodyParser(messageBundle); err != nil {
		return err
	}

	err := utils.Validate(messageBundle)
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

	messageBundle.Message.OwnerID = userId

	messageBundle.Message.Check()

	id, err := messageBundle.Insert(db)
	if err != nil {
		if err.(*pq.Error).Code.Name() == "foreign_key_violation" {
			if err.(*pq.Error).Constraint == "recipient_recipient_id_fkey" {
				return fiber.NewError(fiber.StatusBadRequest, "not valid recipient")
			} else {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("not valid owner %d this should not be possible", userId))
			}
		}
		log.Printf("failed to insert messag: %s\n", err)
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{"id": id})
}

func deleteMessage(c *fiber.Ctx) error {
	messageId, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		return MissingMessageID
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

	message := models.Message{ID: messageId, OwnerID: userId}

	affected, err := message.Delete(db)
	if err != nil {
		log.Printf("failed to delete message: %s\n", err)
		return fiber.ErrInternalServerError
	}
	if affected == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "message does not exits or token is not the owner")
	}
	return c.SendString("message deleted")
}

func getMessages(c *fiber.Ctx) error {
	messageBundle := models.MessageBundle{}

	authorization := c.Get(fiber.HeaderAuthorization)
	userId, err := utils.ValidateAuth(authorization)
	if err != nil {
		return err
	}
	messageBundle.Message.OwnerID = userId
	messageIdStr := c.Query("id")
	if messageIdStr != "" {
		messageId, err := strconv.Atoi(messageIdStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "message id need to a int")
		}

		messageBundle.Message.ID = messageId
		return getMessage(c, &messageBundle)
	}

	db, err := utils.Connection()
	if err != nil {
		log.Printf("failed to connect to db: %s\n", err)
		return fiber.ErrInternalServerError
	}

	messageBundles, err := messageBundle.GetAll(db, true)
	if err != nil {
		log.Printf("failed to get messages %s\n", err)
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{
		"messages": messageBundles,
	})
}

func getMessage(c *fiber.Ctx, messageBundle *models.MessageBundle) error {
	db, err := utils.Connection()
	if err != nil {
		log.Printf("failed to connect to db: %s\n", err)
		return fiber.ErrInternalServerError
	}

	err = messageBundle.Get(db, true)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "message not found")
		}
		log.Printf("failed to get message %s\n", err)
		return fiber.ErrInternalServerError
	}
	return c.JSON(messageBundle)
}

func editMessage(c *fiber.Ctx) error {
	messageId, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		return MissingMessageID
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

	message := models.Message{ID: messageId, OwnerID: userId}

	err = message.Get(db, false)

	if err := c.BodyParser(&message); err != nil {
		return err
	}

	message.Check()

	err = message.Update(db)
	if err != nil {
		log.Printf("failed to update message %s\n", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(message)
}

func Message(app *fiber.App) {
	route := app.Group("/message")

	route.Post("", addMessage)
	route.Delete("", deleteMessage)
	route.Get("", getMessages)
	route.Patch("", editMessage)
}
