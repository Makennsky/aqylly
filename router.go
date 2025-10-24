package aqylly

import (
	"context"
	"net/http"
	"sync"
)

// Router is the main router instance
type Router struct {
	trees      map[string]*node
	middleware []HandlerFunc
	pool       sync.Pool

	// HTTP/2 configuration
	HTTP2Config *HTTP2Config
	EnableHTTP2 bool

	// HTTP/3 configuration (placeholder for future)
	EnableHTTP3 bool

	// NotFound handler
	NotFound HandlerFunc

	// MethodNotAllowed handler
	MethodNotAllowed HandlerFunc

	// Handle OPTIONS requests automatically
	HandleOPTIONS bool

	// Internal HTTP server for graceful shutdown
	server *http.Server
}

// New creates a new router instance
func New() *Router {
	r := &Router{
		trees:         make(map[string]*node),
		HTTP2Config:   DefaultHTTP2Config(),
		EnableHTTP2:   true,  // HTTP/2 enabled by default
		EnableHTTP3:   false, // HTTP/3 disabled by default
		HandleOPTIONS: true,
	}

	r.pool.New = func() interface{} {
		return newContext(nil, nil)
	}

	return r
}

// Default creates a router with default middleware (Logger and Recovery)
func Default() *Router {
	r := New()
	r.Use(Logger(), Recovery())
	return r
}

// Use adds middleware to the router
func (r *Router) Use(middleware ...HandlerFunc) {
	r.middleware = append(r.middleware, middleware...)
}

// addRoute adds a route to the router
func (r *Router) addRoute(method, path string, handler HandlerFunc) {
	if path[0] != '/' {
		panic("path must begin with '/'")
	}

	root := r.trees[method]
	if root == nil {
		root = &node{}
		r.trees[method] = root
	}

	root.addRoute(path, method, handler)
}

// GET registers a GET route
func (r *Router) GET(path string, handler HandlerFunc) {
	r.addRoute(http.MethodGet, path, handler)
}

// POST registers a POST route
func (r *Router) POST(path string, handler HandlerFunc) {
	r.addRoute(http.MethodPost, path, handler)
}

// PUT registers a PUT route
func (r *Router) PUT(path string, handler HandlerFunc) {
	r.addRoute(http.MethodPut, path, handler)
}

// DELETE registers a DELETE route
func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.addRoute(http.MethodDelete, path, handler)
}

// PATCH registers a PATCH route
func (r *Router) PATCH(path string, handler HandlerFunc) {
	r.addRoute(http.MethodPatch, path, handler)
}

// HEAD registers a HEAD route
func (r *Router) HEAD(path string, handler HandlerFunc) {
	r.addRoute(http.MethodHead, path, handler)
}

// OPTIONS registers an OPTIONS route
func (r *Router) OPTIONS(path string, handler HandlerFunc) {
	r.addRoute(http.MethodOptions, path, handler)
}

// Any registers a route for all HTTP methods
func (r *Router) Any(path string, handler HandlerFunc) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
	}

	for _, method := range methods {
		r.addRoute(method, path, handler)
	}
}

// Group creates a new route group
func (r *Router) Group(prefix string, middleware ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		prefix:     prefix,
		parent:     nil,
		router:     r,
		middleware: middleware,
	}
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Get context from pool
	c := r.pool.Get().(*Context)
	c.Writer = w
	c.Request = req
	c.ctx = req.Context()
	c.Params = make(map[string]string)
	c.index = -1
	c.queryCache = nil
	c.statusCode = http.StatusOK

	// Find handler
	method := req.Method
	path := req.URL.Path

	if root := r.trees[method]; root != nil {
		if handler, params := root.getValue(path, method); handler != nil {
			c.Params = params

			// Build handlers chain (middleware + handler)
			c.handlers = make([]HandlerFunc, 0, len(r.middleware)+1)
			c.handlers = append(c.handlers, r.middleware...)
			c.handlers = append(c.handlers, handler)

			// Execute chain
			c.Next()

			// Put context back to pool
			r.pool.Put(c)
			return
		}
	}

	// Handle OPTIONS automatically if enabled
	if method == http.MethodOptions && r.HandleOPTIONS {
		r.handleOPTIONS(c, path)
		r.pool.Put(c)
		return
	}

	// Check if path exists with different method
	for m := range r.trees {
		if m != method {
			if root := r.trees[m]; root != nil {
				if handler, _ := root.getValue(path, m); handler != nil {
					// Method not allowed
					if r.MethodNotAllowed != nil {
						c.handlers = []HandlerFunc{r.MethodNotAllowed}
						c.Next()
					} else {
						http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
					}
					r.pool.Put(c)
					return
				}
			}
		}
	}

	// Not found
	if r.NotFound != nil {
		c.handlers = []HandlerFunc{r.NotFound}
		c.Next()
	} else {
		http.NotFound(w, req)
	}

	r.pool.Put(c)
}

// handleOPTIONS handles OPTIONS requests automatically
func (r *Router) handleOPTIONS(c *Context, path string) {
	allowed := make([]string, 0, 7)

	for method := range r.trees {
		if root := r.trees[method]; root != nil {
			if handler, _ := root.getValue(path, method); handler != nil {
				allowed = append(allowed, method)
			}
		}
	}

	if len(allowed) > 0 {
		c.SetHeader("Allow", joinMethods(allowed))
		c.Status(http.StatusNoContent)
	} else {
		c.Status(http.StatusNotFound)
	}
}

// Run starts the HTTP server
func (r *Router) Run(addr string) error {
	r.server = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return r.server.ListenAndServe()
}

// RunTLS starts the HTTPS server with HTTP/2 enabled by default
func (r *Router) RunTLS(addr, certFile, keyFile string) error {
	r.server = &http.Server{
		Addr:      addr,
		Handler:   r,
		TLSConfig: ConfigureTLSForHTTP2(),
	}

	// Configure HTTP/2
	if r.EnableHTTP2 {
		if err := ConfigureHTTP2Server(r.server, r.HTTP2Config); err != nil {
			return err
		}
	}

	return r.server.ListenAndServeTLS(certFile, keyFile)
}

// RunH2C starts the HTTP/2 Cleartext server (no TLS)
// This is useful for internal microservices that don't need TLS
func (r *Router) RunH2C(addr string) error {
	handler := NewH2CHandler(r, r.HTTP2Config)

	r.server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	return r.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (r *Router) Shutdown(ctx context.Context) error {
	if r.server == nil {
		return nil
	}
	return r.server.Shutdown(ctx)
}

// Helper function to join HTTP methods
func joinMethods(methods []string) string {
	result := ""
	for i, method := range methods {
		if i > 0 {
			result += ", "
		}
		result += method
	}
	return result
}
