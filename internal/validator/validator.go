package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator"
)

var validate = validator.New()

func Validate(v any) error {
	if err := validate.Struct(v); err != nil {
		var errs validator.ValidationErrors
		if ok := AsValidationErrors(err, &errs); ok {
			messages := make([]string, 0, len(errs))
			for _, fe := range errs {
				messages = append(messages, fieldError(fe))
			}
			return fmt.Errorf("%s", strings.Join(messages, "; "))
		}
		return err
	}
	return nil
}

func AsValidationErrors(err error, target *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*target = ve
		return true
	}
	return false
}

func fieldError(fe validator.FieldError) string {
	field := toSnakeCase(fe.Field())
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "len", "numeric":
		if field == "token" {
			return "token must be exactly 6 digits"
		}
		return field + "is invalid"
	default:
		return field + " is invalid"
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r | 0x20)
	}
	return result.String()
}
