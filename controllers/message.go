package controllers

import (
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

func AddMessage(db *gorm.DB, message *models.Message, userId uint) interface{} {
	message.OwnerID = userId
	message.Check()
	db.Clauses(clause.Returning{}).Save(message)
	return fiber.Map{"id": message.ID}
}

func DeleteMessage(db *gorm.DB, messageId int, userId uint) (interface{}, error) {
	result := db.Where("id = ?", messageId).Where("owner_id = ?", userId).Delete(&models.Message{})
	if result.Error != nil {
		log.Printf("failed to delete message: %s\n", result.Error)
		return nil, fiber.ErrInternalServerError
	}
	if result.RowsAffected == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "message does not exits or token is not the owner")
	}
	return fiber.Map{"result": "message deleted"}, nil
}

func GetMessages(db *gorm.DB, userId uint) (interface{}, error) {
	var messages []models.Message
	var messagesFromAssociation []models.Message

	db.Where("owner_id = ?", userId).Find(&messages)

	err := db.Model(&models.Message{}).Where("id = ?", userId).Association("Recipients").Find(&messagesFromAssociation)

	copy(messages[len(messages):], messagesFromAssociation)

	if err != nil {
		log.Printf("failed to get messages %s\n", err)
		return nil, fiber.ErrInternalServerError
	}
	return fiber.Map{"messages": messages}, nil
}

func GetMessage(db *gorm.DB, messageId int, userId uint) (interface{}, error) {
	var message models.Message

	result := db.Where("id = ?", messageId).Where("owner_id = ?", userId).Limit(1).Find(&message)

	if result.Error != nil {
		log.Printf("failed to delete message: %s\n", result.Error)
		return nil, fiber.ErrInternalServerError
	}
	if result.RowsAffected == 0 {
		return nil, fiber.NewError(fiber.StatusNotFound, "message not found")
	}
	return message, nil
}

func EditMessage(db *gorm.DB, updateMessage func(out interface{}) error, messageId int, userId uint) (interface{}, error) {
	var message models.Message

	result := db.Where("id = ?", messageId).Where("owner_id = ?", userId).Limit(1).Find(&message)
	if result.Error != nil {
		log.Printf("failed to delete message: %s\n", result.Error)
		return nil, fiber.ErrInternalServerError
	}
	if result.RowsAffected == 0 {
		return nil, fiber.NewError(fiber.StatusNotFound, "message not found")
	}

	if err := updateMessage(&message); err != nil {
		return nil, err
	}

	message.Check()

	result = db.Save(&message)
	if result.Error != nil {
		log.Printf("failed to update message %s\n", result.Error)
		return nil, fiber.ErrInternalServerError
	}

	return message, nil
}
