package notifier

import (
	"context"
	"github.com/carousell/go-logging"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"strings"
)

const (
	tracerID = "tracerId"
)

// SetTraceId updates the traceID based on context values
func SetTraceId(ctx context.Context) context.Context {
	if GetTraceId(ctx) != "" {
		return ctx
	}
	var traceID string
	if span := stdopentracing.SpanFromContext(ctx); span != nil {
		traceID = span.BaggageItem("trace")
	}
	// if no trace id then create one
	if strings.TrimSpace(traceID) == "" {
		traceID = uuid.NewUUID().String()
	}
	ctx = logging.AddToLogContext(ctx, "trace", traceID)
	return AddToOptions(ctx, tracerID, traceID)
}

// GetTraceId fetches traceID from context
func GetTraceId(ctx context.Context) string {
	if o := FromContext(ctx); o != nil {
		if data, found := o.Get(tracerID); found {
			return data.(string)
		}
	}
	if logCtx := logging.FromContext(ctx); logCtx != nil {
		if data, found := logCtx["trace"]; found {
			traceID := data.(string)
			AddToOptions(ctx, tracerID, traceID)
			return traceID
		}
	}
	return ""
}
