package main

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type (
	ErrorHandlerFunc func(recovery interface{}, c *gin.Context)
)

func ErrorHandler(handlers ...ErrorHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			rec := recover()
			for _, handler := range handlers {
				handler(rec, c)
			}

			if rec != nil || len(c.Errors) > 0 {
				c.Abort()
			}
		}()

		c.Next()
	}
}

func ErrorResponseHandler() ErrorHandlerFunc {
	return func(recovery interface{}, c *gin.Context) {
		publicErrors := c.Errors.ByType(gin.ErrorTypePublic)
		privateLen := len(c.Errors.ByType(gin.ErrorTypePrivate))
		publicLen := len(publicErrors)

		if privateLen == 0 && publicLen == 0 && recovery == nil {
			return
		}

		messagesLen := publicLen
		if privateLen > 0 || recovery != nil {
			messagesLen++
		}

		messages := make([]string, messagesLen)
		index := 0
		for _, err := range publicErrors {
			messages[index] = err.Error()
			index++
		}

		if privateLen > 0 || recovery != nil {
			messages[index] = "Something went wrong"
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": messages})
	}
}

func ErrorCaptureHandler(client *raven.Client, errorsStacktrace bool) ErrorHandlerFunc {
	return func(recovery interface{}, c *gin.Context) {
		tags := map[string]string{
			"endpoint": c.Request.RequestURI,
		}

		if recovery != nil {
			stacktrace := raven.NewStacktrace(4, 3, nil)
			recStr := fmt.Sprint(recovery)
			err := errors.New(recStr)
			go client.CaptureMessageAndWait(
				recStr,
				tags,
				raven.NewException(err, stacktrace),
				raven.NewHttp(c.Request),
			)
		}

		for _, err := range c.Errors {
			if errorsStacktrace {
				stacktrace := NewRavenStackTrace(client, err.Err, 0)
				go client.CaptureMessageAndWait(
					err.Error(),
					tags,
					raven.NewException(err.Err, stacktrace),
					raven.NewHttp(c.Request),
				)
			} else {
				go client.CaptureErrorAndWait(err.Err, tags)
			}
		}
	}
}

func PanicLogger() ErrorHandlerFunc {
	return func(recovery interface{}, c *gin.Context) {
		if recovery != nil {
			fmt.Printf("===\n%+v\n", recovery)
			debug.PrintStack()
		}
	}
}

func ErrorLogger() ErrorHandlerFunc {
	return func(recovery interface{}, c *gin.Context) {
		for _, err := range c.Errors {
			fmt.Printf("===\n%+v\n", err.Err)
		}
	}
}

func NewRavenStackTrace(client *raven.Client, myerr error, skip int) *raven.Stacktrace {
	st := getErrorStackTraceConverted(myerr, 3, client.IncludePaths())
	if st == nil {
		st = raven.NewStacktrace(skip, 3, client.IncludePaths())
	}
	return st
}

func getErrorStackTraceConverted(err error, context int, appPackagePrefixes []string) *raven.Stacktrace {
	st := getErrorCauseStackTrace(err)
	if st == nil {
		return nil
	}
	return convertStackTrace(st, context, appPackagePrefixes)
}

func getErrorCauseStackTrace(err error) errors.StackTrace {
	// This code is inspired by github.com/pkg/errors.Cause().
	var st errors.StackTrace
	for err != nil {
		s := getErrorStackTrace(err)
		if s != nil {
			st = s
		}
		err = getErrorCause(err)
	}
	return st
}

func convertStackTrace(st errors.StackTrace, context int, appPackagePrefixes []string) *raven.Stacktrace {
	// This code is borrowed from github.com/getsentry/raven-go.NewStacktrace().
	var frames []*raven.StacktraceFrame
	for _, f := range st {
		frame := convertFrame(f, context, appPackagePrefixes)
		if frame != nil {
			frames = append(frames, frame)
		}
	}
	if len(frames) == 0 {
		return nil
	}
	for i, j := 0, len(frames)-1; i < j; i, j = i+1, j-1 {
		frames[i], frames[j] = frames[j], frames[i]
	}
	return &raven.Stacktrace{Frames: frames}
}

func convertFrame(f errors.Frame, context int, appPackagePrefixes []string) *raven.StacktraceFrame {
	// This code is borrowed from github.com/pkg/errors.Frame.
	pc := uintptr(f) - 1
	fn := runtime.FuncForPC(pc)
	var file string
	var line int
	if fn != nil {
		file, line = fn.FileLine(pc)
	} else {
		file = "unknown"
	}
	return raven.NewStacktraceFrame(pc, file, line, context, appPackagePrefixes)
}

func getErrorStackTrace(err error) errors.StackTrace {
	ster, ok := err.(interface {
		StackTrace() errors.StackTrace
	})
	if !ok {
		return nil
	}
	return ster.StackTrace()
}

func getErrorCause(err error) error {
	cer, ok := err.(interface {
		Cause() error
	})
	if !ok {
		return nil
	}
	return cer.Cause()
}

type errorResponse struct {
	Errors []string `json:"errors"`
}

func NotFound(errors ...string) (int, interface{}) {
	return http.StatusNotFound, errorResponse{
		Errors: errors,
	}
}

func BadRequest(errors ...string) (int, interface{}) {
	return http.StatusBadRequest, errorResponse{
		Errors: errors,
	}
}
