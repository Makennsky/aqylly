package aqylly

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Context represents the context of the current HTTP request
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request

	// Context for cancellation, timeouts, and values
	ctx context.Context

	// URL params (/users/:id)
	Params map[string]string

	// Parsed query params
	queryCache url.Values

	// Index for middleware chain
	index int

	// Handlers chain (middleware + final handler)
	handlers []HandlerFunc

	// Status code
	statusCode int
}

// HandlerFunc defines the handler used by middleware and routes
type HandlerFunc func(*Context)

// newContext creates a new Context
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	return &Context{
		Writer:     w,
		Request:    r,
		ctx:        ctx,
		Params:     make(map[string]string),
		index:      -1,
		statusCode: http.StatusOK,
	}
}

// Next executes the next handler in the chain
func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Param returns the value of the URL param
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// Query returns the query param value
func (c *Context) Query(key string) string {
	if c.queryCache == nil {
		c.queryCache = c.Request.URL.Query()
	}
	return c.queryCache.Get(key)
}

// QueryDefault returns the query param value or default if not found
func (c *Context) QueryDefault(key, defaultValue string) string {
	if value := c.Query(key); value != "" {
		return value
	}
	return defaultValue
}

// QueryInt returns the query param as int
func (c *Context) QueryInt(key string) (int, error) {
	value := c.Query(key)
	if value == "" {
		return 0, fmt.Errorf("query param '%s' not found", key)
	}
	return strconv.Atoi(value)
}

// QueryIntDefault returns the query param as int or default if not found
func (c *Context) QueryIntDefault(key string, defaultValue int) int {
	value, err := c.QueryInt(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// QueryBool returns the query param as bool
func (c *Context) QueryBool(key string) (bool, error) {
	value := c.Query(key)
	if value == "" {
		return false, fmt.Errorf("query param '%s' not found", key)
	}
	return strconv.ParseBool(value)
}

// QueryArray returns the query param as array
func (c *Context) QueryArray(key string) []string {
	if c.queryCache == nil {
		c.queryCache = c.Request.URL.Query()
	}
	return c.queryCache[key]
}

// Header returns request header value
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets response header
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// Status sets the HTTP status code
func (c *Context) Status(code int) *Context {
	c.statusCode = code
	c.Writer.WriteHeader(code)
	return c
}

// JSON sends a JSON response
func (c *Context) JSON(code int, obj interface{}) error {
	c.SetHeader("Content-Type", "application/json; charset=utf-8")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	return encoder.Encode(obj)
}

// String sends a plain text response
func (c *Context) String(code int, format string, values ...interface{}) error {
	c.SetHeader("Content-Type", "text/plain; charset=utf-8")
	c.Status(code)
	_, err := fmt.Fprintf(c.Writer, format, values...)
	return err
}

// HTML sends an HTML response
func (c *Context) HTML(code int, html string) error {
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.Status(code)
	_, err := c.Writer.Write([]byte(html))
	return err
}

// Data sends raw bytes
func (c *Context) Data(code int, contentType string, data []byte) error {
	c.SetHeader("Content-Type", contentType)
	c.Status(code)
	_, err := c.Writer.Write(data)
	return err
}

// BindJSON binds request body as JSON
func (c *Context) BindJSON(obj interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	return decoder.Decode(obj)
}

// Body returns the raw request body
func (c *Context) Body() ([]byte, error) {
	return io.ReadAll(c.Request.Body)
}

// Method returns the HTTP method
func (c *Context) Method() string {
	return c.Request.Method
}

// Path returns the request path
func (c *Context) Path() string {
	return c.Request.URL.Path
}

// FullPath returns the full request URL
func (c *Context) FullPath() string {
	return c.Request.URL.String()
}

// ClientIP returns the client IP address
func (c *Context) ClientIP() string {
	// Check X-Forwarded-For header
	if ip := c.Header("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, get the first one
		if index := strings.Index(ip, ","); index != -1 {
			return strings.TrimSpace(ip[:index])
		}
		return ip
	}

	// Check X-Real-IP header
	if ip := c.Header("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to RemoteAddr
	if index := strings.LastIndex(c.Request.RemoteAddr, ":"); index != -1 {
		return c.Request.RemoteAddr[:index]
	}

	return c.Request.RemoteAddr
}

// ContentType returns the Content-Type header
func (c *Context) ContentType() string {
	return c.Header("Content-Type")
}

// IsJSON checks if the request content type is JSON
func (c *Context) IsJSON() bool {
	return strings.Contains(c.ContentType(), "application/json")
}

// IsXML checks if the request content type is XML
func (c *Context) IsXML() bool {
	contentType := c.ContentType()
	return strings.Contains(contentType, "application/xml") ||
		strings.Contains(contentType, "text/xml")
}

// IsForm checks if the request content type is form
func (c *Context) IsForm() bool {
	contentType := c.ContentType()
	return strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data")
}

// FormValue returns the form value
func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

// PostForm returns the post form value
func (c *Context) PostForm(key string) string {
	return c.Request.PostFormValue(key)
}

// Redirect redirects to the given URL
func (c *Context) Redirect(code int, location string) {
	http.Redirect(c.Writer, c.Request, location, code)
}

// Cookie returns the cookie value
func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// SetCookie sets a cookie
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	}
	http.SetCookie(c.Writer, cookie)
}

// Context returns the underlying context.Context
func (c *Context) Context() context.Context {
	return c.ctx
}

// WithContext replaces the context
func (c *Context) WithContext(ctx context.Context) {
	c.ctx = ctx
	c.Request = c.Request.WithContext(ctx)
}

// Get retrieves data from context (for middleware communication)
func (c *Context) Get(key string) (interface{}, bool) {
	val := c.ctx.Value(key)
	if val != nil {
		return val, true
	}
	return nil, false
}

// Set stores data in context (for middleware communication)
func (c *Context) Set(key string, value interface{}) {
	c.ctx = context.WithValue(c.ctx, key, value)
	c.Request = c.Request.WithContext(c.ctx)
}

// WithTimeout sets a timeout for the context
func (c *Context) WithTimeout(timeout time.Duration) (context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(c.ctx, timeout)
	c.WithContext(ctx)
	return cancel, nil
}

// WithDeadline sets a deadline for the context
func (c *Context) WithDeadline(deadline time.Time) (context.CancelFunc, error) {
	ctx, cancel := context.WithDeadline(c.ctx, deadline)
	c.WithContext(ctx)
	return cancel, nil
}

// WithCancel creates a cancellable context
func (c *Context) WithCancel() context.CancelFunc {
	ctx, cancel := context.WithCancel(c.ctx)
	c.WithContext(ctx)
	return cancel
}

// Deadline returns the deadline from the context
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

// Done returns the done channel from the context
func (c *Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Err returns the error from the context
func (c *Context) Err() error {
	return c.ctx.Err()
}

// Abort prevents pending handlers from being called
func (c *Context) Abort() {
	c.index = len(c.handlers)
}

// AbortWithStatus aborts and sends status code
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

// AbortWithJSON aborts and sends JSON response
func (c *Context) AbortWithJSON(code int, obj interface{}) {
	c.Abort()
	c.JSON(code, obj)
}

// Error sends an error response
func (c *Context) Error(code int, err error) error {
	return c.JSON(code, map[string]string{
		"error": err.Error(),
	})
}

// Push initiates an HTTP/2 server push
// This allows the server to send resources to the client before they are requested
func (c *Context) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := c.Writer.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}
