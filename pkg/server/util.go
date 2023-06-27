package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var validate *validator.Validate

func Decode(r *http.Request, val any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(val); err != nil {
		return errors.WithMessage(err, "failed to decode request")
	}

	if err := validate.Struct(val); err != nil {
		return errors.WithMessage(err, "failed to validate request")
	}

	return nil
}

// Respond convert a Go value to JSON and sends it to the client.
func Respond(c *gin.Context, data any, statusCode int) {
	// check if the data is an error
	if err, ok := data.(error); ok && err != nil {
		c.PureJSON(statusCode, err.Error())
		return
	}

	// if there's no payload to marshal, set the status code of the response and return
	if statusCode == http.StatusNoContent {
		c.Status(statusCode)
		return
	}

	// respond with pretty JSON
	c.PureJSON(statusCode, data)
}

// LoggingRespondError sends an error response back to the client as a safe error
func LoggingRespondError(c *gin.Context, err error, statusCode int) {
	logrus.WithError(err).Error()
	Respond(c, err, statusCode)
}

// LoggingRespondErrMsg sends an error response back to the client as a safe error from a msg
func LoggingRespondErrMsg(c *gin.Context, errMsg string, statusCode int) {
	LoggingRespondError(c, errors.New(errMsg), statusCode)
}

// LoggingRespondErrWithMsg sends an error response back to the client as a safe error from an error and msg
func LoggingRespondErrWithMsg(c *gin.Context, err error, errMsg string, statusCode int) {
	LoggingRespondError(c, errors.Wrap(err, errMsg), statusCode)
}
