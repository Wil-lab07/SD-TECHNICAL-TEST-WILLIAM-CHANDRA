package validation

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func Validate(i any) error {
	return validate.Struct(i)
}
