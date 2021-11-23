package hermesErrors

import "github.com/gofiber/fiber/v2"

func MessageDoesNotExits() *BaseError {
	return &BaseError{
		fiberError: fiber.NewError(fiber.StatusNotFound, "message does not exits or token is not the owner"),
	}
}

func RecipientDoesNotExits() *BaseError {
	return &BaseError{
		fiberError: fiber.NewError(fiber.StatusBadRequest, "message recipient does not exits"),
	}
}

func OwnerDoesNotExits() *BaseError {
	return &BaseError{
		fiberError: fiber.NewError(fiber.StatusBadRequest, "message owner does not exits this should not be possible"),
	}
}
