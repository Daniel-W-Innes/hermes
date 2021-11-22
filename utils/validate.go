package utils

import (
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/go-playground/validator"
)

// Validate user input, checks if the inputted model passes the 'validate' tags
func Validate(s interface{}) hermesErrors.HermesError {
	var errors []*hermesErrors.ValidatorError
	validate := validator.New()
	errs := validate.Struct(s)
	// build ValidatorError from the errors outputted by validate
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			errors = append(errors, &hermesErrors.ValidatorError{
				Field: err.StructNamespace(),
				Tag:   err.Tag(),
				Value: err.Param(),
			})
		}
		// build ValidatorErrors into a hermesError
		return hermesErrors.FailedValidation(errors)
	}
	return nil
}
