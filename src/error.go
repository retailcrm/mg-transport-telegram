package main

import (
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func BadRequest(error string) (int, interface{}) {
	return http.StatusBadRequest, ErrorResponse{
		Error: getLocalizedMessage(error),
	}
}
