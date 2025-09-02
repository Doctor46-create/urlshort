package model

import (
	dictionary "github.com/Doctor46-create/urlshort/internal/utils"
	"github.com/go-playground/validator/v10"
)

func InputJSONValidate(inputJSON any) error {
	err := dictionary.Validate.Struct(inputJSON)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return err
		}
	}
	return nil
}
