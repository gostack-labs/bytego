package bytego

import (
	"net/http"
	"path"
	"sync"
)

type Router interface {
	GET(path string, handlers ...HandlerFunc) Router
	POST(path string, handlers ...HandlerFunc) Router
	PUT(path string, handlers ...HandlerFunc) Router
	DELETE(path string, handlers ...HandlerFunc) Router
	HEAD(path string, handlers ...HandlerFunc) Router
	PATCH(path string, handlers ...HandlerFunc) Router
	OPTIONS(path string, handlers ...HandlerFunc) Router
	TRACE(path string, handlers ...HandlerFunc) Router
	Handle(method string, path string, handlers ...HandlerFunc) Router
	Any(path string, handlers ...HandlerFunc) Router
	Static(relativePath, root string) Router
	StaticFS(relativePath string, fsys http.FileSystem) Router
	StaticFile(relativePath, filePath string) Router
	Group(relativePath string, handlers ...HandlerFunc) Router
	Use(middlewares ...HandlerFunc)
}

func newRouter() *router {
	r := &router{
		basePath: "/",
	}
	return r
}

type router struct {
	basePath           string
	paramsPool         sync.Pool
	trees              map[string]*node
	maxParams          uint16
	handlers           []HandlerFunc
	noRouteHandlers    []HandlerFunc
	allNoRouteHandlers []HandlerFunc
}

func (r *router) GET(path string, handlers ...HandlerFunc) Router {
	return r.Handle(http.MethodGet, path, handlers...)
}

func (r *router) POST(path string, handlers ...HandlerFunc) Router {
	return r.Handle(http.MethodPost, path, handlers...)
}

func (r *router) PUT(path string, handlers ...HandlerFunc) Router {
	return r.Handle(http.MethodPut, path, handlers...)
}

func (r *router) DELETE(path string, handlers ...HandlerFunc) Router {
	return r.Handle(http.MethodDelete, path, handlers...)
}

func (r *router) HEAD(path string, handlers ...HandlerFunc) Router {
	return r.Handle(http.MethodHead, path, handlers...)
}

func (r *router) PATCH(path string, handlers ...HandlerFunc) Router {
	return r.Handle(http.MethodPatch, path, handlers...)
}

func (r *router) OPTIONS(path string, handlers ...HandlerFunc) Router {
	return r.Handle(http.MethodOptions, path, handlers...)
}

func (r *router) TRACE(path string, handlers ...HandlerFunc) Router {
	return r.Handle(http.MethodTrace, path, handlers...)
}

func (r *router) Any(path string, handlers ...HandlerFunc) Router {
	for _, method := range anyMethods {
		r.Handle(method, path, handlers...)
	}
	return r
}

func (r *router) Handle(method string, path string, handlers ...HandlerFunc) Router {
	path = joinPath(r.basePath, path)
	return r.add(method, path, handlers...)
}

func (r *router) Group(relativePath string, handlers ...HandlerFunc) Router {
	return &Group{
		basePath: joinPath(r.basePath, relativePath),
		route:    r,
		handlers: combineHandlers(r.handlers, handlers),
	}
}

func (r *router) Static(relativePath, root string) Router {
	return r.StaticFS(relativePath, http.Dir(root))
}

func (r *router) StaticFS(relativePath string, fsys http.FileSystem) Router {
	prefix := joinPath(r.basePath, relativePath)
	fileServer := http.StripPrefix(prefix, http.FileServer(fsys))
	handlerFunc := func(c *Ctx) error {
		// filepath := c.Param("filepath")
		fileServer.ServeHTTP(c.Response, c.Request)
		return nil
	}
	return r.GET(path.Join(relativePath, "/*filepath"), handlerFunc)
}

func (r *router) StaticFile(relativePath, filePath string) Router {
	handlerFunc := func(c *Ctx) error {
		http.ServeFile(c.Response, c.Request, filePath)
		return nil
	}
	return r.GET(relativePath, handlerFunc)
}

func (r *router) Use(middlewares ...HandlerFunc) {
	r.handlers = append(r.handlers, middlewares...)
	r.rebuild404Handlers()
}

func (r *router) noRoute(handlers ...HandlerFunc) {
	r.noRouteHandlers = handlers
	r.rebuild404Handlers()
}

func (r *router) rebuild404Handlers() {
	r.allNoRouteHandlers = combineHandlers(r.handlers, r.noRouteHandlers)
}

func (r *router) add(method, path string, handlers ...HandlerFunc) Router {
	varsCount := uint16(0)

	if method == "" {
		panic("method must not be empty")
	}
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	if handlers == nil {
		panic("handle must not be nil")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	root.addRoute(path, combineHandlers(r.handlers, handlers)...)

	// Update maxParams
	if paramsCount := countParams(path); paramsCount+varsCount > r.maxParams {
		r.maxParams = paramsCount + varsCount
	}

	// Lazy-init paramsPool alloc func
	if r.paramsPool.New == nil && r.maxParams > 0 {
		r.paramsPool.New = func() interface{} {
			ps := make(Params, 0, r.maxParams)
			return &ps
		}
	}
	return nil
}

func (r *router) getParams() *Params {
	ps, _ := r.paramsPool.Get().(*Params)
	*ps = (*ps)[0:0] // reset slice
	return ps
}

func (r *router) putParams(ps *Params) {
	if ps != nil {
		r.paramsPool.Put(ps)
	}
}

func combineHandlers(handlers1, handlers2 []HandlerFunc) []HandlerFunc {
	size := len(handlers1) + len(handlers2)
	mergedHandlers := make([]HandlerFunc, size)
	copy(mergedHandlers, handlers1)
	copy(mergedHandlers[len(handlers1):], handlers2)
	return mergedHandlers
}
