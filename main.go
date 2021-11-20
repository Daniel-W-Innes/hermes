package main

import (
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/Daniel-W-Innes/hermes/routes"
	"github.com/Daniel-W-Innes/hermes/utils"
	"github.com/gofiber/fiber/v2"
	"log"
)

func health(c *fiber.Ctx) error {
	return c.SendString("ok")
}

func initDB(config *models.DBConfig) error {
	db, hermesError := utils.Connection(config)
	if hermesError != nil {
		return hermesError
	}
	err := db.AutoMigrate(&models.Message{}, &models.User{})
	if err != nil {
		return err
	}
	return nil
}

func getApp() *fiber.App {
	app := fiber.New()

	app.Get("/", health)

	routes.User(app)
	routes.Message(app)
	return app
}

func main() {
	config, err := models.GetConfig()
	if err != nil {
		log.Panic(err)
	}

	err = initDB(&config.DBConfig)
	if err != nil {
		log.Panic(err)
	}

	err = getApp().Listen(":8080")
	if err != nil {
		log.Panic(err)
	}
}
