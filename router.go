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
	GET(path string, handler HandlerFunc) Router
	POST(path string, handler HandlerFunc) Router
	PUT(path string, handler HandlerFunc) Router
	DELETE(path string, handler HandlerFunc) Router
	HEAD(path string, handler HandlerFunc) Router
	PATCH(path string, handler HandlerFunc) Router
	OPTIONS(path string, handler HandlerFunc) Router
	Handle(method string, path string, handler HandlerFunc) Router
	Any(path string, handler HandlerFunc) Router
	Group(relativePath string, handlers ...HandlerFunc) *Group
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

func newRouter() Router {
	return &route{}
}

type route struct {
	basePath   string
	paramsPool sync.Pool
	trees      map[string]*node
	maxParams  uint16
}

func (r *route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if root := r.trees[req.Method]; root != nil {
		if handle, ps, _ := root.getValue(path, r.getParams); handle != nil {
			ctx := &Ctx{
				path:    path,
				Request: req,
				Writer:  w,
			}
			if ps != nil {
				handle(ctx)
				r.putParams(ps)
			} else {
				handle(ctx)
			}
			return
		}
	}

	http.NotFound(w, req)
}

func (r *route) GET(path string, handler HandlerFunc) Router {
	return r.add(http.MethodGet, path, handler)
}

func (r *route) POST(path string, handler HandlerFunc) Router {
	return r.add(http.MethodPost, path, handler)
}

func (r *route) PUT(path string, handler HandlerFunc) Router {
	return r.add(http.MethodPut, path, handler)
}

func (r *route) DELETE(path string, handler HandlerFunc) Router {
	return r.add(http.MethodDelete, path, handler)
}

func (r *route) HEAD(path string, handler HandlerFunc) Router {
	return r.add(http.MethodHead, path, handler)
}

func (r *route) PATCH(path string, handler HandlerFunc) Router {
	return r.add(http.MethodPatch, path, handler)
}

func (r *route) OPTIONS(path string, handler HandlerFunc) Router {
	return r.add(http.MethodOptions, path, handler)
}

func (r *route) Any(path string, handler HandlerFunc) Router {
	for _, method := range anyMethods {
		r.add(method, path, handler)
	}
	return r
}

func (r *route) Handle(method string, path string, handler HandlerFunc) Router {
	return r.add(method, path, handler)
}

func (r *route) Group(relativePath string, handlers ...HandlerFunc) *Group {
	return &Group{
		basePath: joinPath(r.basePath, relativePath),
		route:    r,
	}
}

func (r *route) add(method, path string, handler HandlerFunc) Router {
	varsCount := uint16(0)

	if method == "" {
		panic("method must not be empty")
	}
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	if handler == nil {
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

	root.addRoute(path, handler)

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
