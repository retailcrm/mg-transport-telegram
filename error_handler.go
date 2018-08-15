package main

import (
	"fmt"
	"net/http"
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
			logger.Error(recovery)
			debug.PrintStack()
		}
	}
}

func ErrorLogger() ErrorHandlerFunc {
	return func(recovery interface{}, c *gin.Context) {
		for _, err := range c.Errors {
			logger.Error(err.Err)
		}
	}
}
