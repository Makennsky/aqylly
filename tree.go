package aqylly

// nodeType represents the type of route node
type nodeType uint8

const (
	static nodeType = iota // default
	root
	param   // :param
	catchAll // *param
)

// node represents a node in the radix tree
type node struct {
	path      string
	indices   string
	wildChild bool
	nType     nodeType
	priority  uint32
	children  []*node
	handlers  map[string]HandlerFunc
	params    []string
}

// addRoute adds a route to the tree
func (n *node) addRoute(path string, method string, handler HandlerFunc) {
	n.priority++

	// Empty tree
	if len(n.path) == 0 && len(n.children) == 0 {
		n.insertChild(path, method, handler)
		n.nType = root
		return
	}

walk:
	for {
		// Find the longest common prefix
		i := longestCommonPrefix(path, n.path)

		// Split edge
		if i < len(n.path) {
			child := node{
				path:      n.path[i:],
				wildChild: n.wildChild,
				nType:     static,
				indices:   n.indices,
				children:  n.children,
				handlers:  n.handlers,
				priority:  n.priority - 1,
				params:    n.params,
			}

			n.children = []*node{&child}
			n.indices = string([]byte{n.path[i]})
			n.path = path[:i]
			n.handlers = nil
			n.wildChild = false
			n.params = nil
		}

		// Make new node a child of this node
		if i < len(path) {
			path = path[i:]
			c := path[0]

			// Check if a child with the next path byte exists
			for i, maxIdx := 0, len(n.indices); i < maxIdx; i++ {
				if c == n.indices[i] {
					i = n.incrementChildPrio(i)
					n = n.children[i]
					continue walk
				}
			}

			// Otherwise insert it
			if c != ':' && c != '*' {
				n.indices += string([]byte{c})
				child := &node{}
				n.children = append(n.children, child)
				n.incrementChildPrio(len(n.indices) - 1)
				n = child
			}

			n.insertChild(path, method, handler)
			return
		}

		// Otherwise add handler to current node
		if n.handlers == nil {
			n.handlers = make(map[string]HandlerFunc)
		}
		n.handlers[method] = handler
		return
	}
}

// insertChild inserts a child node
func (n *node) insertChild(path, method string, handler HandlerFunc) {
	for {
		// Find prefix until first wildcard
		wildcard, i, valid := findWildcard(path)
		if i < 0 { // No wildcard found
			break
		}

		// The wildcard name must not contain ':' and '*'
		if !valid {
			panic("only one wildcard per path segment is allowed")
		}

		// Check if the wildcard has a name
		if len(wildcard) < 2 {
			panic("wildcards must be named with a non-empty name")
		}

		// Check if this node has existing children which would be unreachable
		if len(n.children) > 0 {
			panic("wildcard segment conflicts with existing children")
		}

		if wildcard[0] == ':' { // param
			if i > 0 {
				// Insert prefix before the current wildcard
				n.path = path[:i]
				path = path[i:]
			}

			n.wildChild = true
			child := &node{
				nType: param,
				path:  wildcard,
			}
			n.children = []*node{child}
			n = child
			n.priority++

			// If the path doesn't end with the wildcard, then there
			// will be another non-wildcard subpath starting with '/'
			if len(wildcard) < len(path) {
				path = path[len(wildcard):]
				child := &node{
					priority: 1,
				}
				n.children = []*node{child}
				n = child
				continue
			}

			// Otherwise we're done. Insert the handler
			if n.handlers == nil {
				n.handlers = make(map[string]HandlerFunc)
			}
			n.handlers[method] = handler
			n.params = append(n.params, wildcard[1:]) // Remove ':'
			return

		} else { // catchAll
			if i+len(wildcard) != len(path) {
				panic("catch-all routes are only allowed at the end of the path")
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				panic("catch-all conflicts with existing handle for the path segment root")
			}

			// Currently fixed width 1 for '/'
			i--
			if path[i] != '/' {
				panic("no / before catch-all")
			}

			n.path = path[:i]

			// First node: catchAll node with empty path
			child := &node{
				wildChild: true,
				nType:     catchAll,
			}
			n.children = []*node{child}
			n.indices = string(path[i])
			n = child
			n.priority++

			// Second node: node holding the variable
			child = &node{
				path:     path[i:],
				nType:    catchAll,
				handlers: make(map[string]HandlerFunc),
				priority: 1,
			}
			child.handlers[method] = handler
			child.params = append(child.params, wildcard[1:]) // Remove '*'
			n.children = []*node{child}

			return
		}
	}

	// If no wildcard was found, simply insert the path and handler
	n.path = path
	if n.handlers == nil {
		n.handlers = make(map[string]HandlerFunc)
	}
	n.handlers[method] = handler
}

// getValue returns the handler and params for a given path
func (n *node) getValue(path, method string) (handler HandlerFunc, params map[string]string) {
	params = make(map[string]string)

walk:
	for {
		prefix := n.path
		if len(path) > len(prefix) {
			if path[:len(prefix)] == prefix {
				path = path[len(prefix):]

				// If this node does not have a wildcard child,
				// we can just look up the next child node and continue
				if !n.wildChild {
					c := path[0]
					for i, maxIdx := 0, len(n.indices); i < maxIdx; i++ {
						if c == n.indices[i] {
							n = n.children[i]
							continue walk
						}
					}

					// Nothing found
					return nil, nil
				}

				// Handle wildcard child
				n = n.children[0]
				switch n.nType {
				case param:
					// Find param end
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// Save param value
					if len(n.path) > 0 {
						params[n.path[1:]] = path[:end]
					}

					// We need to go deeper!
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						// ... but we can't
						return nil, nil
					}

					if handler := n.handlers[method]; handler != nil {
						return handler, params
					}

					if len(n.children) == 0 {
						return nil, nil
					}

					// Check for handle on the current node
					n = n.children[0]
					if handler := n.handlers[method]; handler != nil {
						return handler, params
					}

					return nil, nil

				case catchAll:
					// Save param value
					if len(n.path) > 1 {
						params[n.params[0]] = path
					}

					if handler := n.handlers[method]; handler != nil {
						return handler, params
					}
					return nil, nil

				default:
					panic("invalid node type")
				}
			}
		} else if path == prefix {
			// We should have reached the node containing the handler
			if handler := n.handlers[method]; handler != nil {
				return handler, params
			}

			return nil, nil
		}

		// Nothing found
		return nil, nil
	}
}

// incrementChildPrio increments the priority of a child and reorders if necessary
func (n *node) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	// Adjust position (move to front)
	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {
		// Swap node positions
		cs[newPos-1], cs[newPos] = cs[newPos], cs[newPos-1]
	}

	// Build new index char string
	if newPos != pos {
		n.indices = n.indices[:newPos] + // Unchanged prefix
			n.indices[pos:pos+1] + // The index char we move
			n.indices[newPos:pos] + n.indices[pos+1:] // Rest without char at 'pos'
	}

	return newPos
}

// Helper functions

// longestCommonPrefix finds the longest common prefix
func longestCommonPrefix(a, b string) int {
	i := 0
	maxLen := minInt(len(a), len(b))
	for i < maxLen && a[i] == b[i] {
		i++
	}
	return i
}

// findWildcard finds wildcard segments
func findWildcard(path string) (wildcard string, i int, valid bool) {
	// Find start
	for start, c := range []byte(path) {
		// A wildcard starts with ':' (param) or '*' (catch-all)
		if c != ':' && c != '*' {
			continue
		}

		// Find end and check for invalid characters
		valid = true
		for end, c := range []byte(path[start+1:]) {
			switch c {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	return "", -1, false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
