package main

import (
	"reflect"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v8"
)

var regCommandName = regexp.MustCompile(`https://?[\da-z.-]+\.(retailcrm\.(ru|pro|es)|ecomlogic\.com)`)

func setValidation() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("validatecrmurl", validateCrmURL)
	}
}

func validateCrmURL(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	return regCommandName.Match([]byte(field.Interface().(string)))
}
