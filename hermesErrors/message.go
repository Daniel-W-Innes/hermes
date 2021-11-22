package hermesErrors

import "github.com/gofiber/fiber/v2"

func MessageDoesNotExits() *BaseError {
	return &BaseError{
		fiberError: fiber.NewError(fiber.StatusBadRequest, "message does not exits or token is not the owner"),
	}
}
