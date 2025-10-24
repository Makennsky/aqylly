package aqylly

// RouterGroup is used to group routes with a common prefix and middleware
type RouterGroup struct {
	prefix     string
	parent     *RouterGroup
	router     *Router
	middleware []HandlerFunc
}

// Group creates a new nested route group
func (g *RouterGroup) Group(prefix string, middleware ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		prefix:     g.prefix + prefix,
		parent:     g,
		router:     g.router,
		middleware: append(g.combineMiddleware(), middleware...),
	}
}

// Use adds middleware to the group
func (g *RouterGroup) Use(middleware ...HandlerFunc) {
	g.middleware = append(g.middleware, middleware...)
}

// combineMiddleware combines parent and current middleware
func (g *RouterGroup) combineMiddleware() []HandlerFunc {
	if g.parent != nil {
		return append(g.parent.combineMiddleware(), g.middleware...)
	}
	return append([]HandlerFunc{}, g.middleware...)
}

// handle registers a route with group middleware
func (g *RouterGroup) handle(method, path string, handler HandlerFunc) {
	fullPath := g.prefix + path

	// Combine group middleware with handler
	finalHandler := func(c *Context) {
		// Inject group middleware before the handler
		groupMiddleware := g.combineMiddleware()

		// Save original handlers
		originalHandlers := c.handlers
		originalIndex := c.index

		// Create new handlers chain with group middleware
		newHandlers := make([]HandlerFunc, 0, len(groupMiddleware)+1)
		newHandlers = append(newHandlers, groupMiddleware...)
		newHandlers = append(newHandlers, handler)

		// Replace handlers
		c.handlers = newHandlers
		c.index = -1

		// Execute the chain
		c.Next()

		// Restore original handlers
		c.handlers = originalHandlers
		c.index = originalIndex
	}

	g.router.addRoute(method, fullPath, finalHandler)
}

// GET registers a GET route in the group
func (g *RouterGroup) GET(path string, handler HandlerFunc) {
	g.handle("GET", path, handler)
}

// POST registers a POST route in the group
func (g *RouterGroup) POST(path string, handler HandlerFunc) {
	g.handle("POST", path, handler)
}

// PUT registers a PUT route in the group
func (g *RouterGroup) PUT(path string, handler HandlerFunc) {
	g.handle("PUT", path, handler)
}

// DELETE registers a DELETE route in the group
func (g *RouterGroup) DELETE(path string, handler HandlerFunc) {
	g.handle("DELETE", path, handler)
}

// PATCH registers a PATCH route in the group
func (g *RouterGroup) PATCH(path string, handler HandlerFunc) {
	g.handle("PATCH", path, handler)
}

// HEAD registers a HEAD route in the group
func (g *RouterGroup) HEAD(path string, handler HandlerFunc) {
	g.handle("HEAD", path, handler)
}

// OPTIONS registers an OPTIONS route in the group
func (g *RouterGroup) OPTIONS(path string, handler HandlerFunc) {
	g.handle("OPTIONS", path, handler)
}

// Any registers a route for all HTTP methods in the group
func (g *RouterGroup) Any(path string, handler HandlerFunc) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		g.handle(method, path, handler)
	}
}
