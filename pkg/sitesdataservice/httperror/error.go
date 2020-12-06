package httperror

import (
	"context"
	"net/http"
	"strings"

	"github.com/valyala/fasthttp"
)

// Error is a service error.
type Error struct {
	Code    int
	Message string
	Log     string
}

// Error returns a text message corresponding to the given error.
func (e *Error) Error() string {
	return e.Log
}

// StatusCode returns an HTTP status code corresponding to the given error.
func (e *Error) StatusCode() int {
	return e.Code
}

// ErrorProcessor ...
type ErrorProcessor interface {
	Encode(ctx context.Context, r *fasthttp.Response, err error)
	Decode(r *fasthttp.Response) error
}

type errorProcessor struct {
	defaultCode    int
	defaultMessage string
}

// Encode writes a svc error to the given http.ResponseWriter.
func (e *errorProcessor) Encode(ctx context.Context, r *fasthttp.Response, err error) {
	code := e.defaultCode
	message := e.defaultMessage
	if err, ok := err.(*Error); ok {
		if err.Code != e.defaultCode {
			code = err.Code
			message = err.Message
		}
	}
	r.SetStatusCode(code)
	r.SetBodyString(message)
}

// Decode reads a Service error from the given *http.Response.
func (e *errorProcessor) Decode(r *fasthttp.Response) error {
	msgBytes := r.Body()
	msg := strings.TrimSpace(string(msgBytes))
	if msg == "" {
		msg = http.StatusText(r.StatusCode())
	}
	return &Error{
		Code:    r.StatusCode(),
		Message: msg,
	}
}

// NewErrorProcessor ...
func NewErrorProcessor(defaultCode int, defaultMessage string) ErrorProcessor {
	return &errorProcessor{
		defaultCode:    defaultCode,
		defaultMessage: defaultMessage,
	}
}

// ErrorCreator ...
type ErrorCreator func(status int, message string, log string) error

// NewError ...
func NewError(status int, message string, log string) error {
	return &Error{
		Code:    status,
		Message: message,
		Log:     log,
	}
}
