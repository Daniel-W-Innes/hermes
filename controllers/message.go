package controllers

import (
	"fmt"
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AddMessage add a new message and check all word play types
func AddMessage(db *gorm.DB, message *models.Message, userId uint) (interface{}, hermesErrors.HermesError) {
	// ensure the OwnerID matches the userId from the auth
	message.OwnerID = userId

	// check all word play types
	message.Check()

	// create message in db
	result := db.Clauses(clause.Returning{}).Create(message)
	if result.Error != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to add message %s\n", result.Error))
	}

	// return message id for get, delete and edit calls
	return fiber.Map{"id": message.ID}, nil
}

// DeleteMessage delete a message by id
func DeleteMessage(db *gorm.DB, messageId int, userId uint) (fiber.Map, hermesErrors.HermesError) {
	// delete the message and specify owner_id prevent from deleting other users message
	result := db.Where("id = ?", messageId).Where("owner_id = ?", userId).Delete(&models.Message{})
	if result.Error != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to delete message: %s\n", result.Error))
	}

	if result.RowsAffected == 0 {
		return nil, hermesErrors.MessageDoesNotExits()
	}
	return fiber.Map{"result": "message deleted"}, nil
}

// GetMessages get all
func GetMessages(db *gorm.DB, userId uint) (fiber.Map, hermesErrors.HermesError) {
	var messages []models.Message
	var messagesFromAssociation []models.Message

	result := db.Where("owner_id = ?", userId).Find(&messages)
	if result.Error != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to get owned messages %s\n", result.Error))
	}

	err := db.Model(&models.Message{}).Where("id = ?", userId).Association("Recipients").Find(&messagesFromAssociation)
	if err != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to get messages by recipients %s\n", result.Error))
	}
	copy(messages[len(messages):], messagesFromAssociation)

	return fiber.Map{"messages": messages}, nil
}

func GetMessage(db *gorm.DB, messageId int, userId uint) (*models.Message, hermesErrors.HermesError) {
	var message models.Message

	result := db.Where("id = ?", messageId).Where("owner_id = ?", userId).Limit(1).Find(&message)

	if result.Error != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to delete message: %s\n", result.Error))
	}
	if result.RowsAffected == 0 {
		return nil, hermesErrors.MessageDoesNotExits()
	}
	return &message, nil
}

func EditMessage(db *gorm.DB, updateMessage func(out interface{}) error, messageId int, userId uint) (*models.Message, hermesErrors.HermesError) {
	var message models.Message

	result := db.Where("id = ?", messageId).Where("owner_id = ?", userId).Limit(1).Find(&message)
	if result.Error != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to delete message: %s\n", result.Error))
	}
	if result.RowsAffected == 0 {
		return nil, hermesErrors.MessageDoesNotExits()
	}

	if err := updateMessage(&message); err != nil {
		return nil, hermesErrors.UnprocessableEntity(fmt.Sprintf("failed to update message with user input: %s\n", err))
	}

	message.Check()

	result = db.Save(&message)
	if result.Error != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to update message %s\n", result.Error))
	}

	return &message, nil
}
