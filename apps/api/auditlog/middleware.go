package auditlog

import (
	"bytes"
	"context"
	"github.com/gin-contrib/requestid"
	"github.com/juju/errors"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

// Clock provides time.
type Clock interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

// ErrorHandler handles an error.
type ErrorHandler interface {
	Handle(err error)
	HandleContext(ctx context.Context, err error)
}

// NoopErrorHandler is an error handler that discards every error.
type NoopErrorHandler struct{}

func (NoopErrorHandler) Handle(_ error)                           {}
func (NoopErrorHandler) HandleContext(_ context.Context, _ error) {}

// Stores combine multiple stores into a single instance.
type Stores []Store

func combineErrors(errs ...error) (combined error) {
	for _, err := range errs {
		combined = errors.Wrap(combined, err)
	}
	return
}

func (d Stores) Store(entry Entry) error {
	var errs []error

	for _, store := range d {
		errs = append(errs, store.Store(entry))
	}

	return combineErrors(errs...)
}

// Entry holds all information related to an API call event.
type Entry struct {
	Time      time.Time
	RequestID string
	ContextID string
	OrgID     string
	UserID    string
	HTTP      HTTPEntry
}

type SerializedError []byte

func (e SerializedError) String() string {
	return string(e)
}

// HTTPEntry contains details related to an HTTP call for an audit log entry.
type HTTPEntry struct {
	ClientIP     string
	UserAgent    string
	Method       string
	Path         string
	RequestBody  string
	StatusCode   int
	ResponseTime int
	ResponseSize int
	Errors       []SerializedError
}

type realClock struct{}

func (realClock) Now() time.Time                  { return time.Now() }
func (realClock) Since(t time.Time) time.Duration { return time.Since(t) }

// Store saves audit log entries.
type Store interface {
	// Store saves an audit log entry.
	Store(entry Entry) error
}

// Option configures an audit log middleware.
type Option interface {
	// apply is unexported,
	// so only the current package can implement this interface.
	apply(o *middlewareOptions)
}

// In the future, sensitivePaths and userIDExtractor might be replaced by request matchers and propagators/decorators
// respectively to generalize them for multiple use cases, but for now this solution (borrowed from the previous one)
// should be fine.
type middlewareOptions struct {
	clock           Clock
	sensitivePaths  []*regexp.Regexp
	userIDExtractor func(c *gin.Context) (contextID, ownerID, userID string)
	errorHandler    ErrorHandler
	ignorePaths     []*regexp.Regexp
}

type optionFunc func(o *middlewareOptions)

func (fn optionFunc) apply(o *middlewareOptions) {
	fn(o)
}

// WithClock sets the clock in an audit log middleware.
func WithClock(clock Clock) Option {
	return optionFunc(func(o *middlewareOptions) {
		o.clock = clock
	})
}

// WithIgnorePaths marks API call paths as ignored, causing the log entry to omit the request body.
func WithIgnorePaths(ignorePaths []*regexp.Regexp) Option {
	return optionFunc(func(o *middlewareOptions) {
		o.ignorePaths = ignorePaths
	})
}

// WithSensitivePaths marks API call paths as sensitive, causing the log entry to omit the request body.
func WithSensitivePaths(sensitivePaths []*regexp.Regexp) Option {
	return optionFunc(func(o *middlewareOptions) {
		o.sensitivePaths = sensitivePaths
	})
}

// WithErrorHandler sets the clock in an audit log middleware.
func WithErrorHandler(errorHandler ErrorHandler) Option {
	return optionFunc(func(o *middlewareOptions) {
		o.errorHandler = errorHandler
	})
}

// WithIDExtractor sets the function that extracts the user ID from the request.
func WithIDExtractor(userIDExtractor func(c *gin.Context) (contextID, ownerID, userID string)) Option {
	return optionFunc(func(o *middlewareOptions) {
		o.userIDExtractor = userIDExtractor
	})
}

// Middleware returns a new HTTP middleware that records audit log entries.
func Middleware(store Store, opts ...Option) gin.HandlerFunc {
	options := middlewareOptions{
		clock:           realClock{},
		userIDExtractor: func(c *gin.Context) (string, string, string) { return "", "", "" },
		errorHandler:    NoopErrorHandler{},
	}

	for _, opt := range opts {
		opt.apply(&options)
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		for _, ignorePath := range options.ignorePaths {
			if ignorePath.MatchString(path) {
				c.Next() // process request without logging
				return
			}
		}

		entry := Entry{
			Time:      options.clock.Now(),
			RequestID: requestid.Get(c),
			HTTP: HTTPEntry{
				ClientIP:  c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Method:    c.Request.Method,
				Path:      path,
			},
		}

		var sensitiveCall bool

		// Determine if this call contains sensitive information in its request body.
		for _, r := range options.sensitivePaths {
			if r.MatchString(c.Request.URL.Path) {
				sensitiveCall = true
				break
			}
		}

		// Only override the request body if there is actually one and it doesn't contain sensitive information.
		saveBody := c.Request.Body != nil && !sensitiveCall

		var buf bytes.Buffer

		if saveBody {
			// This should be ok, because the server keeps a reference to the original body,
			// so it can close the original request itself.
			c.Request.Body = io.NopCloser(io.TeeReader(c.Request.Body, &buf))
		}

		c.Next() // process request

		contextID, ownerID, userID := options.userIDExtractor(c)
		entry.ContextID = contextID
		entry.OrgID = ownerID
		entry.UserID = userID

		// Consider making this configurable if you need to log unauthorized requests,
		// but keep in mind that in case of a public installation it's a potential DoS attack vector.
		if c.Writer.Status() == http.StatusUnauthorized {
			return
		}

		entry.HTTP.StatusCode = c.Writer.Status()
		entry.HTTP.ResponseSize = c.Writer.Size()
		entry.HTTP.ResponseTime = int(options.clock.Since(entry.Time).Milliseconds())

		if saveBody {
			// Make sure everything is read from the body.
			_, err := io.ReadAll(c.Request.Body)
			if err != nil && err != io.EOF {
				options.errorHandler.HandleContext(c.Request.Context(), errors.Maskf(err, "Failed to read response body"))
			}

			entry.HTTP.RequestBody = string(buf.Bytes())
		}

		if c.IsAborted() {
			for _, e := range c.Errors {
				_e, _ := e.MarshalJSON()

				entry.HTTP.Errors = append(entry.HTTP.Errors, _e)
			}
		}

		err := store.Store(entry)
		if err != nil {
			options.errorHandler.HandleContext(c.Request.Context(), errors.Maskf(err, "Failed to save audit log"))
		}
	}
}
