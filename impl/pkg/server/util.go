package server

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	validate *validator.Validate
)

func init() {
	// Instantiate validator.
	validate = validator.New()

	// Use JSON tag names for errors instead of Go struct field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})
}

// Respond convert a Go value to JSON and sends it to the client.
func Respond(c *gin.Context, data any, statusCode int) {
	// check if the data is an error
	if err, ok := data.(error); ok && err != nil {
		c.PureJSON(statusCode, err.Error())
		return
	}

	// if there's no payload to marshal, set the status code of the response and return
	if statusCode == http.StatusNoContent || data == nil {
		c.Status(statusCode)
		return
	}

	// respond with pretty JSON
	c.PureJSON(statusCode, data)
}

// ResponseStatus sends a response with a status code and no body.
func ResponseStatus(c *gin.Context, statusCode int) {
	c.Status(statusCode)
}

// RespondBytes sends a byte array to the client.
func RespondBytes(c *gin.Context, data []byte, statusCode int) {
	// if there's no payload to marshal, set the status code of the response and return
	if statusCode == http.StatusNoContent || data == nil {
		c.Status(statusCode)
		return
	}

	// respond with an octet stream
	c.Data(statusCode, "application/octet-stream", data)
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

// GetParam is a utility to get a path parameter from context, nil if not found
func GetParam(c *gin.Context, param string) *string {
	got := c.Param(param)
	if got == "" {
		return nil
	}
	// remove leading slash, which is a quirk of gin
	if got[0] == '/' {
		got = got[1:]
	}
	return &got
}

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
	})
}
