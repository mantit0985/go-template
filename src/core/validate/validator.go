package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a domain-specific validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("field %s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	msgs := make([]string, 0, len(ve))
	for _, e := range ve {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

// ValidatorAPI defines the interface for validation.
type ValidatorAPI interface {
	Validate(s any) error
}

// validatorImpl wraps the go-playground/validator.Validate instance.
type validatorImpl struct {
	validate *validator.Validate
}

// NewValidator creates a new Validator and returns it as a ValidatorAPI.
func NewValidator() ValidatorAPI {
	v := validator.New()

	// RegisterTagNameFunc allows the validator to use the "json" tag
	// for field names in validation error messages.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &validatorImpl{
		validate: v,
	}
}

// Validate performs validation on the provided struct and wraps errors.
func (v *validatorImpl) Validate(s any) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	if ve, ok := err.(validator.ValidationErrors); ok {
		var res ValidationErrors
		for _, fe := range ve {
			res = append(res, ValidationError{
				Field:   fe.Field(),
				Message: fmt.Sprintf("failed on the %s tag", fe.Tag()),
			})
		}
		return res
	}

	return fmt.Errorf("validation error: %w", err)
}
