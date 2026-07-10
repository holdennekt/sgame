package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/backend/internal/service"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

// NotificationController receives raw-bucket upload notifications (GCS
// Pub/Sub push or Minio webhook); any non-2xx response triggers redelivery.
type NotificationController struct {
	processor *service.AttachmentProcessorService
	token     string
}

func NewNotificationController(processor *service.AttachmentProcessorService, token string) *NotificationController {
	return &NotificationController{processor, token}
}

func (c *NotificationController) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/notifications/attachments", c.handle)
}

type pubSubPush struct {
	Message struct {
		Data       []byte            `json:"data"`
		Attributes map[string]string `json:"attributes"`
	} `json:"message"`
}

type gcsObject struct {
	Name string `json:"name"`
}

type minioWebhookEvent struct {
	EventName string `json:"EventName"`
	Records   []struct {
		EventName string `json:"eventName"`
		S3        struct {
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

// An empty result with nil error means a valid but irrelevant event to ack.
func parseNotificationKeys(body []byte) ([]string, error) {
	var push pubSubPush
	if err := json.Unmarshal(body, &push); err == nil && len(push.Message.Data) > 0 {
		if et, ok := push.Message.Attributes["eventType"]; ok && et != "OBJECT_FINALIZE" {
			return nil, nil
		}
		var obj gcsObject
		if err := json.Unmarshal(push.Message.Data, &obj); err != nil {
			return nil, fmt.Errorf("invalid GCS object payload: %w", err)
		}
		if obj.Name == "" {
			return nil, fmt.Errorf("GCS object payload has no name")
		}
		return []string{obj.Name}, nil
	}

	var event minioWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("unrecognized notification payload: %w", err)
	}
	var keys []string
	for _, record := range event.Records {
		eventName := record.EventName
		if eventName == "" {
			eventName = event.EventName
		}
		if !strings.Contains(eventName, "ObjectCreated") {
			continue
		}
		key, err := url.QueryUnescape(record.S3.Object.Key)
		if err != nil {
			return nil, fmt.Errorf("invalid object key %q: %w", record.S3.Object.Key, err)
		}
		keys = append(keys, key)
	}
	if len(event.Records) == 0 {
		return nil, fmt.Errorf("notification carries no records")
	}
	return keys, nil
}

func (c *NotificationController) authorized(ctx *gin.Context) bool {
	if c.token == "" {
		return true
	}
	auth := ctx.GetHeader("Authorization")
	if strings.TrimPrefix(auth, "Bearer ") == c.token || auth == c.token {
		return true
	}
	return ctx.Query("token") == c.token
}

func (c *NotificationController) handle(ctx *gin.Context) {
	if !c.authorized(ctx) {
		_ = ctx.Error(custerr.NewUnauthorizedErr("invalid notification token"))
		return
	}

	body, err := ctx.GetRawData()
	if err != nil {
		_ = ctx.Error(custerr.NewBadRequestErr(err.Error()))
		return
	}

	keys, err := parseNotificationKeys(body)
	if err != nil {
		_ = ctx.Error(custerr.NewBadRequestErr(err.Error()))
		return
	}

	for _, key := range keys {
		if err := c.processor.Process(ctx, key); err != nil {
			_ = ctx.Error(err)
			return
		}
	}

	ctx.Status(http.StatusNoContent)
}
