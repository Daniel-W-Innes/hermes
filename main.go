package main

import (
	"github.com/Daniel-W-Innes/hermes/routes"
	"github.com/Daniel-W-Innes/hermes/utils"
	"github.com/gofiber/fiber/v2"
	"log"
)

func health(c *fiber.Ctx) error {
	return c.SendString("ok")
}

func initDB() error {
	db, err := utils.Connection()
	if err != nil {
		return err
	}
	tx := db.MustBegin()
	tx.MustExec("CREATE TABLE IF NOT EXISTS app_user(id SERIAL PRIMARY KEY NOT NULL, username VARCHAR UNIQUE NOT NULL, password_key BYTEA NOT NULL )")
	tx.MustExec("CREATE TABLE IF NOT EXISTS message(id SERIAL PRIMARY KEY NOT NULL, owner_id INT REFERENCES app_user(id), text TEXT NOT NULL, palindrome BOOL NOT NULL)")
	tx.MustExec("CREATE TABLE IF NOT EXISTS recipient(message_id INT REFERENCES message(id), recipient_id INT REFERENCES app_user(id))")
	err = tx.Commit()
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
	err := initDB()
	if err != nil {
		log.Panic(err)
	}

	err = getApp().Listen(":8080")
	if err != nil {
		log.Panic(err)
	}
}
