package controllers

import "github.com/go-playground/validator/v10"

type ErrorMsg struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func getErrorMsg(fe validator.FieldError) string {
	return fe.Error()
}
