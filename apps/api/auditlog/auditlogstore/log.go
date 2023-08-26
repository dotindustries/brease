package auditlogstore

import (
	"go.dot.industries/brease/auditlog"
	"go.uber.org/zap"
)

// LogConfig configures a standard output audit log store.
type LogConfig struct {
	Verbosity int
	Fields    []string
}

// Logger is the fundamental interface for all log operations.
type Logger interface {
	// Info logs an info event.
	Info(msg string, fields ...zap.Field)
}

// NewLog returns a standard output audit log store.
func NewLog(config LogConfig, logger Logger) auditlog.Store {
	return log{
		config: config,
		logger: logger,
	}
}

type log struct {
	config LogConfig
	logger Logger
}

func (d log) Store(entry auditlog.Entry) error {
	var fields []zap.Field

	if len(d.config.Fields) > 0 {
		fields = appendFields(fields, entry, d.config.Fields)
	} else {
		if d.config.Verbosity == 0 {
			return nil
		}

		if d.config.Verbosity >= 1 {
			fields = appendFields(fields, entry, []string{"timestamp", "requestID", "userID"})
		}

		if d.config.Verbosity >= 2 {
			fields = appendFields(fields, entry, []string{"http.method", "http.path", "http.clientIP"})
		}

		if d.config.Verbosity >= 3 {
			fields = appendFields(fields, entry, []string{"http.userAgent", "http.statusCode", "http.responseTime", "http.responseSize"})
		}

		if d.config.Verbosity >= 4 {
			fields = appendFields(fields, entry, []string{"http.requestBody", "http.errors"})
		}
	}

	d.logger.Info("Audit log event", fields...)

	return nil
}

func appendFields(fields []zap.Field, entry auditlog.Entry, dataFields []string) []zap.Field {
	for _, field := range dataFields {
		var add *zap.Field
		switch field {
		case "timestamp":
			f := zap.Time("timestamp", entry.Time)
			add = &f
		case "requestID":
			f := zap.String("requestID", entry.RequestID)
			add = &f
		case "userID":
			f := zap.Any("userID", entry.UserID)
			add = &f
		case "http.method":
			f := zap.String("http.method", entry.HTTP.Method)
			add = &f
		case "http.path":
			f := zap.String("http.method", entry.HTTP.Path)
			add = &f
		case "http.clientIP":
			f := zap.String("http.clientIP", entry.HTTP.ClientIP)
			add = &f
		case "http.userAgent":
			f := zap.String("http.userAgent", entry.HTTP.UserAgent)
			add = &f
		case "http.statusCode":
			f := zap.Int("http.statusCode", entry.HTTP.StatusCode)
			add = &f
		case "http.responseTime":
			f := zap.Int("http.responseTime", entry.HTTP.ResponseTime)
			add = &f
		case "http.responseSize":
			f := zap.Int("http.responseSize", entry.HTTP.ResponseSize)
			add = &f
		case "http.requestBody":
			f := zap.String("http.requestBody", entry.HTTP.RequestBody)
			add = &f
		case "http.errors":
			if len(entry.HTTP.Errors) > 0 {
				f := zap.Stringers("http.method", entry.HTTP.Errors)
				add = &f
			}
		}
		if add != nil {
			fields = append(fields, *add)
		}
	}
	return fields
}
