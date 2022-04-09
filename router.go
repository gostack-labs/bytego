package bytego

import (
	"net/http"
	"sync"
)

var (
	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

type Router interface {
	GET(path string, handlers ...HandlerFunc) Router
	POST(path string, handlers ...HandlerFunc) Router
	PUT(path string, handlers ...HandlerFunc) Router
	DELETE(path string, handlers ...HandlerFunc) Router
	HEAD(path string, handlers ...HandlerFunc) Router
	PATCH(path string, handlers ...HandlerFunc) Router
	OPTIONS(path string, handlers ...HandlerFunc) Router
	Handle(method string, path string, handlers ...HandlerFunc) Router
	Any(path string, handlers ...HandlerFunc) Router
	Group(relativePath string, handlers ...HandlerFunc) *Group
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

func newRouter() Router {
	route := &route{
		basePath: "/",
	}
	route.pool.New = func() interface{} {
		return &Ctx{}
	}
	return route
}

type route struct {
	basePath   string
	paramsPool sync.Pool
	trees      map[string]*node
	maxParams  uint16
	pool       sync.Pool
}

func (r *route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if root := r.trees[req.Method]; root != nil {
		if handlers, ps, _ := root.getValue(path, r.getParams); handlers != nil {
			ctx := r.pool.Get().(*Ctx)
			ctx.reset()
			ctx.path = path
			ctx.Request = req
			ctx.Writer = w
			ctx.handlers = handlers

			if ps != nil {
				ctx.Next()
				r.putParams(ps)
			} else {
				ctx.Next()
			}
			r.pool.Put(ctx)
			return
		}
	}

	http.NotFound(w, req)
}

func (r *route) GET(path string, handlers ...HandlerFunc) Router {
	return r.add(http.MethodGet, path, handlers...)
}

func (r *route) POST(path string, handlers ...HandlerFunc) Router {
	return r.add(http.MethodPost, path, handlers...)
}

func (r *route) PUT(path string, handlers ...HandlerFunc) Router {
	return r.add(http.MethodPut, path, handlers...)
}

func (r *route) DELETE(path string, handlers ...HandlerFunc) Router {
	return r.add(http.MethodDelete, path, handlers...)
}

func (r *route) HEAD(path string, handlers ...HandlerFunc) Router {
	return r.add(http.MethodHead, path, handlers...)
}

func (r *route) PATCH(path string, handlers ...HandlerFunc) Router {
	return r.add(http.MethodPatch, path, handlers...)
}

func (r *route) OPTIONS(path string, handlers ...HandlerFunc) Router {
	return r.add(http.MethodOptions, path, handlers...)
}

func (r *route) Any(path string, handlers ...HandlerFunc) Router {
	for _, method := range anyMethods {
		r.add(method, path, handlers...)
	}
	return r
}

func (r *route) Handle(method string, path string, handlers ...HandlerFunc) Router {
	return r.add(method, path, handlers...)
}

func (r *route) Group(relativePath string, handlers ...HandlerFunc) *Group {
	return &Group{
		basePath: joinPath(r.basePath, relativePath),
		route:    r,
	}
}

func (r *route) add(method, path string, handlers ...HandlerFunc) Router {
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

	root.addRoute(path, handlers...)

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

func (r *route) getParams() *Params {
	ps, _ := r.paramsPool.Get().(*Params)
	*ps = (*ps)[0:0] // reset slice
	return ps
}
func (r *route) putParams(ps *Params) {
	if ps != nil {
		r.paramsPool.Put(ps)
	}
}
