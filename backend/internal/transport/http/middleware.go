package http

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

func baseKind(t reflect.Type) reflect.Kind {
	if t.Kind() == reflect.Pointer {
		return t.Elem().Kind()
	}
	return t.Kind()
}

func messageForTag(fe validator.FieldError) string {
	param := fe.Param()
	switch fe.Tag() {
	case "required":
		return "is required"
	case "required_without":
		return fmt.Sprintf("is required when %s is not provided", strings.ToLower(param))
	case "excluded_with":
		return fmt.Sprintf("cannot be used together with %s", strings.ToLower(param))
	case "min":
		switch baseKind(fe.Type()) {
		case reflect.String:
			return fmt.Sprintf("must be at least %s characters", param)
		case reflect.Int:
			return fmt.Sprintf("must be at least %s", param)
		case reflect.Slice:
			return fmt.Sprintf("must have at least %s items", param)
		}
	case "max":
		switch baseKind(fe.Type()) {
		case reflect.String:
			return fmt.Sprintf("must be at most %s characters", param)
		case reflect.Int:
			return fmt.Sprintf("must be at most %s", param)
		case reflect.Slice:
			return fmt.Sprintf("must have at most %s items", param)
		}
	case "oneof":
		return fmt.Sprintf("must be one of: %s", param)
	case "unique":
		return fmt.Sprintf("must have unique %s", strings.ToLower(param))
	case "same_length":
		return "all categories must have the same number of questions"
	}
	return fmt.Sprintf("failed validation (%s)", fe.Tag())
}

func pathFromNamespace(ns string) string {
	if _, after, ok := strings.Cut(ns, "."); ok {
		return after
	}
	return ns
}

func ErrorMiddleware(ctx *gin.Context) {
	ctx.Next()

	errs := ctx.Errors
	if len(errs) == 0 {
		return
	}
	switch err := errs[0].Err.(type) {
	case validator.ValidationErrors:
		validationErrs := make([]dto.ValidationError, len(err))
		for i, fe := range err {
			validationErrs[i] = dto.ValidationError{
				Path:    pathFromNamespace(fe.Namespace()),
				Message: messageForTag(fe),
			}
		}
		ctx.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{Errors: validationErrs})
	case custerr.BadRequestErr:
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	case custerr.UnauthorizedErr:
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
	case custerr.ForbiddenErr:
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{Error: err.Error()})
	case custerr.NotFoundErr:
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	case custerr.ConflictErr:
		ctx.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}
}

func RequestIDMiddleware(ctx *gin.Context) {
	requestId := ctx.GetHeader("X-Request-ID")
	if requestId == "" {
		requestId = uuid.NewString()
	}
	ctx.Set("requestId", requestId)
	ctx.Header("X-Request-ID", requestId)
	ctx.Next()
}

func LoggingMiddleware(ctx *gin.Context) {
	start := time.Now()

	var requestBodyStr string
	contentType := ctx.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/") {
		requestBodyStr = "[multipart form data]"
	} else if ctx.Request.Body != nil {
		requestBody, _ := io.ReadAll(ctx.Request.Body)
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		requestBodyStr = truncate(string(requestBody), 512)
	}

	respBody := &bytes.Buffer{}
	writer := &bodyWriter{body: respBody, ResponseWriter: ctx.Writer}
	ctx.Writer = writer

	ctx.Next()

	slog.Info("request",
		"request_id", ctx.GetString("requestId"),
		"method", ctx.Request.Method,
		"path", ctx.Request.URL.Path,
		"status", ctx.Writer.Status(),
		"duration_ms", time.Since(start).Milliseconds(),
		"client_ip", ctx.ClientIP(),
		"server_ip", GetServerIP(),
		"request_body", requestBodyStr,
		"response_body", truncate(respBody.String(), 512),
	)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + fmt.Sprintf("... [%d bytes truncated]", len(s)-max)
}

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func GetServerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}
