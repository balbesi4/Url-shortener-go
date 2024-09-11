package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{Status: StatusOK}
}

func Error(message string) Response {
	return Response{
		Status: StatusError,
		Error:  message,
	}
}

func Validation(errs validator.ValidationErrors) Response {
	var errorMessages []string
	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errorMessages = append(errorMessages, fmt.Sprintf("field %s is required", err.Field()))
		case "url":
			errorMessages = append(errorMessages, fmt.Sprintf("field %s is not a valid url", err.Field()))
		default:
			errorMessages = append(errorMessages, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errorMessages, ", "),
	}
}
