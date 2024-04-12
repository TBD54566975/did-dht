package util

import (
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// TraceHook is a logrus hook that adds trace information to log entries
type TraceHook struct{}

func (h *TraceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TraceHook) Fire(entry *logrus.Entry) error {
	ctx := entry.Context
	if ctx == nil {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return nil
	}

	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	entry.Data["trace_id"] = traceID
	entry.Data["span_id"] = spanID

	return nil
}
