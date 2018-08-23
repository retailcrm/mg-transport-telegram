package main

import (
	"reflect"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v8"
)

func setValidation() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("validatecrmurl", validateCrmURL)
	}
}

func validateCrmURL(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	regCommandName := regexp.MustCompile(`https://?[\da-z.-]+\.(retailcrm\.(ru|pro)|ecomlogic\.com)`)

	return regCommandName.Match([]byte(field.Interface().(string)))
}
