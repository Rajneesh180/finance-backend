package api

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func ValidationErrors(err error) string {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}

	var msgs []string
	for _, e := range errs {
		field := strings.ToLower(e.Field())
		switch e.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", field))
		case "email":
			msgs = append(msgs, fmt.Sprintf("%s must be a valid email", field))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s characters", field, e.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s characters", field, e.Param()))
		case "oneof":
			msgs = append(msgs, fmt.Sprintf("%s must be one of: %s", field, e.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s failed %s validation", field, e.Tag()))
		}
	}
	return strings.Join(msgs, "; ")
}
