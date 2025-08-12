package http

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/holdennekt/sgame/pkg/custerr"
)

func messageForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Namespace())
	case "min":
		switch fe.Type().Kind() {
		case reflect.String:
			return fmt.Sprintf("%s must be at least %s characters long", fe.Namespace(), fe.Param())
		case reflect.Int:
			return fmt.Sprintf("%s must be at least %s", fe.Namespace(), fe.Param())
		case reflect.Slice:
			return fmt.Sprintf("%s must have at least %s items", fe.Namespace(), fe.Param())
		}
	case "max":
		switch fe.Type().Kind() {
		case reflect.String:
			return fmt.Sprintf("%s must be at most %s characters long", fe.Namespace(), fe.Param())
		case reflect.Int:
			return fmt.Sprintf("%s must be at most %s", fe.Namespace(), fe.Param())
		case reflect.Slice:
			return fmt.Sprintf("%s must have at most %s items", fe.Namespace(), fe.Param())
		}
	case "url":
		return fmt.Sprintf("%s must be a URL", fe.Namespace())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Namespace(), fe.Param())
	case "unique":
		return fmt.Sprintf("%s must have unique %s values", fe.Namespace(), fe.Param())
	case "same_length":
		return fmt.Sprintf("%s must have same %s length", fe.Namespace(), fe.Param())
	}
	return fe.Error()
}

func ErrorMiddleware(ctx *gin.Context) {
	ctx.Next()

	errs := ctx.Errors
	if len(errs) == 0 {
		return
	}
	switch err := errs[0].Err.(type) {
	case validator.ValidationErrors:
		errs := make([]string, len(err))
		for i, fieldErr := range err {
			errs[i] = messageForTag(fieldErr)
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": strings.Join(errs, ", ")})
	case custerr.BadRequestErr:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case custerr.UnauthorizedErr:
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case custerr.ForbiddenErr:
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case custerr.NotFoundErr:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case custerr.ConflictErr:
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
