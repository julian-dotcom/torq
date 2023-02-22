package server_errors

import (
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const (
	JsonParseError = "Parsing JSON"
)

type ErrorCodeOrDescription struct {
	Code        string            `json:"code"`
	Description string            `json:"description"`
	Attributes  map[string]string `json:"attributes"`
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

func (se *ServerError) AddFieldErrorCode(field string, fieldErrorCode string, attributes map[string]string) {
	if se.Errors.Fields == nil {
		se.Errors.Fields = make(map[string][]ErrorCodeOrDescription)
	}
	if _, exists := se.Errors.Fields[field]; !exists {
		se.Errors.Fields[field] = []ErrorCodeOrDescription{}
	}
	ecod := ErrorCodeOrDescription{Code: fieldErrorCode, Attributes: attributes}
	se.Errors.Fields[field] = append(se.Errors.Fields[field], ecod)
}

func (se *ServerError) AddServerErrorCode(serverErrorCode string, attributes map[string]string) {
	ecod := ErrorCodeOrDescription{Code: serverErrorCode, Attributes: attributes}
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

func SingleServerErrorCode(serverErrorCode string, attributes map[string]string) *ServerError {
	serverError := &ServerError{}
	serverError.AddServerErrorCode(serverErrorCode, attributes)
	return serverError
}

func SingleFieldError(field string, fieldError string) *ServerError {
	serverError := &ServerError{}
	serverError.AddFieldError(field, fieldError)
	return serverError
}

func SingleFieldErrorCode(field string, fieldErrorCode string, attributes map[string]string) *ServerError {
	serverError := &ServerError{}
	serverError.AddFieldErrorCode(field, fieldErrorCode, attributes)
	return serverError
}

func LogAndSendServerError(c *gin.Context, err error) {
	log.Error().Err(err).Send()
	c.JSON(http.StatusInternalServerError, SingleServerError(err.Error()))
}

func LogAndSendServerErrorCode(c *gin.Context, err error, code string, attributes map[string]string) {
	log.Error().Err(err).Send()
	c.JSON(http.StatusInternalServerError, SingleServerErrorCode(code, attributes))
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
