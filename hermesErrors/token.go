package hermesErrors

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

type ValidatorError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

func FailedValidation(errors []*ValidatorError) *BaseError {
	bytes, err := json.Marshal(errors)
	if err != nil {
		return InternalServerError(fmt.Sprintf("failed to marshal validator errors: %s", err))
	}
	return &BaseError{fiberError: fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("user input failed validation %s", bytes))}
}

func UnexpectedSigningMethod(alg interface{}) *BaseError {
	return &BaseError{fiberError: fiber.NewError(fiber.StatusUnauthorized, fmt.Sprintf("unexpected signing method: %v", alg))}
}

func NotValidToken() *BaseError {
	return &BaseError{fiberError: fiber.NewError(fiber.StatusUnauthorized, "token is not valid")}
}

func MissingBearer() *BaseError {
	return &BaseError{fiberError: fiber.NewError(fiber.StatusUnauthorized, "missing bearer in header")}
}
