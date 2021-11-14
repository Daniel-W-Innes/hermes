package utils

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"log"
)

type ValidatorError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

func Validate(s interface{}) error {
	var errors []*ValidatorError
	validate := validator.New()
	errs := validate.Struct(s)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			errors = append(errors, &ValidatorError{
				Field: err.StructNamespace(),
				Tag:   err.Tag(),
				Value: err.Param(),
			})
		}
		bytes, err := json.Marshal(errors)
		if err != nil {
			log.Printf("failed to marshal validator errors: %s", err)
			return fiber.ErrInternalServerError
		}
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("user input failed validation %s", bytes))
	}
	return nil
}
