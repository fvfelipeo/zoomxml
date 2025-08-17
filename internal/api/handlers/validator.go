package handlers

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Usar nome do campo JSON em vez do nome da struct
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// validateStruct valida uma estrutura usando as tags de validação
func validateStruct(s interface{}) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		tag := err.Tag()
		
		switch tag {
		case "required":
			errors[field] = field + " is required"
		case "email":
			errors[field] = field + " must be a valid email"
		case "min":
			errors[field] = field + " must be at least " + err.Param() + " characters"
		case "max":
			errors[field] = field + " must be at most " + err.Param() + " characters"
		case "oneof":
			errors[field] = field + " must be one of: " + err.Param()
		default:
			errors[field] = field + " is invalid"
		}
	}

	return errors
}
