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

	//check if no row were removed
	if result.RowsAffected == 0 {
		return nil, hermesErrors.MessageDoesNotExits()
	}
	return fiber.Map{"result": "message deleted"}, nil
}

// GetMessages get all messages owned by or sent to the user
func GetMessages(db *gorm.DB, userId uint) (fiber.Map, hermesErrors.HermesError) {
	var messages []models.Message

	//get messages from db where user is owner or a recipient
	result := db.Joins("LEFT JOIN recipients r ON messages.id = r.message_id").Where("r.user_id = ?", userId).Or("owner_id = ?", userId).Find(&messages)
	if result.Error != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to get messages %s\n", result.Error))
	}

	return fiber.Map{"messages": messages}, nil
}

// GetMessage get a messages owned by user
func GetMessage(db *gorm.DB, messageId int, userId uint) (*models.Message, hermesErrors.HermesError) {
	var message models.Message

	//get message from db where user is owner or a recipient
	result := db.Joins("LEFT JOIN recipients r ON messages.id = r.message_id").Where("id = ?", messageId).Where(db.Where("r.user_id = ?", userId).Or("owner_id = ?", userId)).Limit(1).Find(&message)
	if result.Error != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to get message: %s\n", result.Error))
	}

	//if message was found, note messages that the user does not have permissions to see will result in this as well
	if result.RowsAffected == 0 {
		return nil, hermesErrors.MessageDoesNotExits()
	}
	return &message, nil
}

// EditMessage update message with user input only if the message is owned by the user
func EditMessage(db *gorm.DB, updateMessage func(out interface{}) error, messageId int, userId uint) (*models.Message, hermesErrors.HermesError) {
	var message models.Message

	//get message from db where user is owner
	result := db.Where("id = ?", messageId).Where("owner_id = ?", userId).Limit(1).Find(&message)
	if result.Error != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to delete message: %s\n", result.Error))
	}
	if result.RowsAffected == 0 {
		return nil, hermesErrors.MessageDoesNotExits()
	}

	//update message using callback to isolate api
	if err := updateMessage(&message); err != nil {
		return nil, hermesErrors.UnprocessableEntity(fmt.Sprintf("failed to update message with user input: %s\n", err))
	}

	//redo check in case text is changed
	message.Check()

	//save changes to the db
	result = db.Save(&message)
	if result.Error != nil {
		return nil, hermesErrors.InternalServerError(fmt.Sprintf("failed to update message %s\n", result.Error))
	}

	return &message, nil
}
