package trace

import (
	"fmt"
	"github.com/gin-gonic/gin"
	trace2 "go.opencensus.io/trace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/semconv/v1.17.0/httpconv"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var Tracer = otel.Tracer("gin-server")

type config struct {
	TracerProvider    oteltrace.TracerProvider
	Propagators       propagation.TextMapPropagator
	Filters           []otelgin.Filter
	SpanNameFormatter otelgin.SpanNameFormatter
}

// Option specifies instrumentation configuration options.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// Middleware returns middleware that will trace incoming requests.
// The service parameter should describe the name of the (virtual)
// server handling the request.
// TODO: this is not reporting anything neither to remote nor to stdout
func Middleware(service string, opts ...Option) gin.HandlerFunc {
	cfg := config{}
	for _, opt := range opts {
		opt.apply(&cfg)
	}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	return func(c *gin.Context) {
		for _, f := range cfg.Filters {
			if !f(c.Request) {
				// Serve the request to the next middleware
				// if a filter rejects the request.
				c.Next()
				return
			}
		}
		savedCtx := c.Request.Context()
		defer func() {
			c.Request = c.Request.WithContext(savedCtx)
		}()
		ctx := cfg.Propagators.Extract(savedCtx, propagation.HeaderCarrier(c.Request.Header))
		var attr []trace2.Attribute
		for _, kv := range httpconv.ServerRequest(service, c.Request) {
			attr = append(attr, translateAttribute(kv))
		}

		var spanName string
		if cfg.SpanNameFormatter == nil {
			spanName = c.FullPath()
		} else {
			spanName = cfg.SpanNameFormatter(c.Request)
		}
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", c.Request.Method)
		} else {
			kv := semconv.HTTPRoute(spanName)
			attr = append(attr, trace2.StringAttribute(string(kv.Key), kv.Value.AsString()))
		}
		ctx, span := trace2.StartSpan(ctx, spanName, trace2.WithSpanKind(trace2.SpanKindServer))
		defer span.End()

		span.AddAttributes(attr...)

		// pass the span through the request context
		c.Request = c.Request.WithContext(ctx)

		// serve the request to the next middleware
		c.Next()

		status := c.Writer.Status()

		if status > 0 {
			_, msg := httpconv.ServerStatus(status)
			span.SetStatus(trace2.Status{
				Code:    int32(status),
				Message: msg,
			})
			kv := semconv.HTTPStatusCode(status)
			span.AddAttributes(trace2.Int64Attribute(string(kv.Key), kv.Value.AsInt64()))
		}
		if len(c.Errors) > 0 {
			span.AddAttributes(trace2.StringAttribute("gin.errors", c.Errors.String()))
		}
	}
}

func translateAttribute(kv attribute.KeyValue) trace2.Attribute {
	switch kv.Value.Type() {
	case attribute.STRING:
		return trace2.StringAttribute(string(kv.Key), kv.Value.AsString())
	case attribute.BOOL:
		return trace2.BoolAttribute(string(kv.Key), kv.Value.AsBool())
	case attribute.INT64:
		return trace2.Int64Attribute(string(kv.Key), kv.Value.AsInt64())
	case attribute.FLOAT64:
		return trace2.Float64Attribute(string(kv.Key), kv.Value.AsFloat64())
	}
	// the rest of the slices will arrive as string
	return trace2.StringAttribute(string(kv.Key), kv.Value.Emit())
}
