package hermesErrors

import "github.com/gofiber/fiber/v2"

func BadLogin() *BaseError {
	return &BaseError{
		fiberError: fiber.NewError(fiber.StatusUnauthorized, "username or password is not right"),
	}
}

func UserExists() *BaseError {
	return &BaseError{
		fiberError: fiber.NewError(fiber.StatusBadRequest, "user already exists"),
	}
}
