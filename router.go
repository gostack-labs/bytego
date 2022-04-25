package bytego

import (
	"net/http"
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
	r.pool.New = func() interface{} {
		return &Ctx{app: r.app}
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
	app                *App
	pool               sync.Pool
}

func (r *router) noRoute(handlers ...HandlerFunc) {
	r.noRouteHandlers = handlers
	r.rebuild404Handlers()
}

func (r *router) rebuild404Handlers() {
	r.allNoRouteHandlers = combineHandlers(r.handlers, r.noRouteHandlers)
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := r.pool.Get().(*Ctx)
	ctx.reset()
	ctx.Request = req
	ctx.writer = newResponseWriter(w)
	ctx.Response = ctx.writer
	defer r.pool.Put(ctx)

	path := req.URL.Path
	if root := r.trees[req.Method]; root != nil {
		if value := root.getValue(path, r.getParams); value.handlers != nil {
			ctx.path = path
			ctx.handlers = value.handlers
			ctx.routePath = value.fullPath
			var err error
			if value.params != nil {
				ctx.Params = *value.params
				err = ctx.Next()
				r.putParams(value.params)
			} else {
				err = ctx.Next()
			}
			if err != nil {
				ctx.HandleError(err)
			}
			return
		}
	}
	ctx.handlers = r.allNoRouteHandlers
	r.serveError(ctx, http.StatusNotFound, default404Body)
}

func (r *router) serveError(c *Ctx, code int, defaultMessage []byte) {
	c.writer.status = code
	_ = c.Next() //middlewares
	if c.Response.Committed() {
		return
	}
	c.Status(code)
	_, _ = c.Response.Write(defaultMessage)
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
