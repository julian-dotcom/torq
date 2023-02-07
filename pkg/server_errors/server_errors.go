package server_errors

import (
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

/*
Error structure expected by the front end
{
	"errors": {
		"fields": {
			"email": ["The email field is required", "The email field must be a valid email"],
			"name": ["The email field is required"],
			"age": ["The age field must be a valid number"]
		},
		"server": [
			"Bad request to the server",
			"Email or password wrong"
		]
	}
}
*/

const (
	JsonParseError = "Parsing JSON"
)

type ErrorCodeOrDescription struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type ServerError struct {
	Errors struct {
		Fields map[string][]ErrorCodeOrDescription `json:"fields"`
		Server []ErrorCodeOrDescription            `json:"server"`
	} `json:"errors"`
}

func (se *ServerError) AddFieldError(field string, fieldError string) {
	if se.Errors.Fields == nil {
		se.Errors.Fields = make(map[string][]ErrorCodeOrDescription)
	}
	if _, exists := se.Errors.Fields[field]; !exists {
		se.Errors.Fields[field] = []ErrorCodeOrDescription{}
	}
	ecod := ErrorCodeOrDescription{Description: fieldError}
	se.Errors.Fields[field] = append(se.Errors.Fields[field], ecod)
}

func (se *ServerError) AddFieldErrorCode(field string, fieldErrorCode string) {
	if se.Errors.Fields == nil {
		se.Errors.Fields = make(map[string][]ErrorCodeOrDescription)
	}
	if _, exists := se.Errors.Fields[field]; !exists {
		se.Errors.Fields[field] = []ErrorCodeOrDescription{}
	}
	ecod := ErrorCodeOrDescription{Code: fieldErrorCode}
	se.Errors.Fields[field] = append(se.Errors.Fields[field], ecod)
}

func (se *ServerError) AddServerErrorCode(serverErrorCode string) {
	ecod := ErrorCodeOrDescription{Code: serverErrorCode}
	se.Errors.Server = append(se.Errors.Server, ecod)
}

func (se *ServerError) AddServerError(serverErrorDescription string) {
	ecod := ErrorCodeOrDescription{Description: serverErrorDescription}
	se.Errors.Server = append(se.Errors.Server, ecod)
}

func SingleServerError(serverErrorDescription string) *ServerError {
	serverError := &ServerError{}
	serverError.AddServerError(serverErrorDescription)
	return serverError
}

func SingleServerErrorCode(serverErrorCode string) *ServerError {
	serverError := &ServerError{}
	serverError.AddServerErrorCode(serverErrorCode)
	return serverError
}

func SingleFieldError(field string, fieldError string) *ServerError {
	serverError := &ServerError{}
	serverError.AddFieldError(field, fieldError)
	return serverError
}

func SingleFieldErrorCode(field string, fieldErrorCode string) *ServerError {
	serverError := &ServerError{}
	serverError.AddFieldErrorCode(field, fieldErrorCode)
	return serverError
}

func LogAndSendServerError(c *gin.Context, err error) {
	log.Error().Err(err).Send()
	c.JSON(http.StatusInternalServerError, SingleServerError(err.Error()))
}

func LogAndSendServerErrorCode(c *gin.Context, err error, code string) {
	log.Error().Err(err).Send()
	c.JSON(http.StatusInternalServerError, SingleServerErrorCode(code))
}

func WrapLogAndSendServerError(c *gin.Context, err error, message string) {
	err = errors.Wrap(err, message)
	log.Error().Err(err).Send()
	c.JSON(http.StatusInternalServerError, SingleServerError(err.Error()))
}

func SendBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, SingleServerError(message))
}

func SendBadRequestFieldError(c *gin.Context, se *ServerError) {
	c.JSON(http.StatusBadRequest, se)
}

func SendUnprocessableEntity(c *gin.Context, message string) {
	c.JSON(http.StatusUnprocessableEntity, SingleServerError(message))
}

func SendBadRequestFromError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, SingleServerError(err.Error()))
}

func SendUnprocessableEntityFromError(c *gin.Context, err error) {
	c.JSON(http.StatusUnprocessableEntity, SingleServerError(err.Error()))
}
