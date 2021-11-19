package utils

import (
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/go-playground/validator"
)

func Validate(s interface{}) hermesErrors.HermesError {
	var errors []*hermesErrors.ValidatorError
	validate := validator.New()
	errs := validate.Struct(s)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			errors = append(errors, &hermesErrors.ValidatorError{
				Field: err.StructNamespace(),
				Tag:   err.Tag(),
				Value: err.Param(),
			})
		}
		return hermesErrors.FailedValidation(errors)
	}
	return nil
}
