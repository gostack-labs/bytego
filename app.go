package bytego

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync"
)

type App struct {
	server *http.Server
	route  *router
	Router
	pool         sync.Pool
	errorHandler ErrorHandler
	binder       *binder
	isDebug      bool
	render       Renderer
}

func New() *App {
	r := newRouter()
	a := &App{
		route:        r,
		Router:       r,
		errorHandler: defaultErrorHandler,
		binder:       &binder{},
	}
	a.pool.New = func() interface{} {
		return &Ctx{app: a}
	}
	return a
}

type HandlerFunc func(*Ctx) error
type Map map[string]interface{}
type Validate func(i interface{}) error
type ValidateTranslate func(err error) error
type Renderer interface {
	Render(io.Writer, string, interface{}) error
}

func (a *App) Use(middlewares ...HandlerFunc) {
	a.Router.Use(middlewares...)
}

func (a *App) Handler() http.Handler {
	return a
}

func (a *App) Validator(fc Validate, trans ...ValidateTranslate) {
	if fc == nil {
		return
	}
	a.binder.validate = fc
	if len(trans) > 0 {
		a.binder.validateTranslate = trans[0]
	}
}

func (a *App) ErrorHandler(fc ErrorHandler) {
	if fc == nil {
		return
	}
	a.errorHandler = fc
}

func (a *App) Render(render Renderer) {
	if render == nil {
		return
	}
	a.render = render
}

func (a *App) Run(addr string) error {
	a.server = &http.Server{
		Addr:    addr,
		Handler: a.Handler(),
	}
	return a.server.ListenAndServe()
}

func (a *App) Stop() error {
	return a.server.Shutdown(context.Background())
}

func (a *App) Debug(isDebug bool) {
	a.isDebug = isDebug
}

func (a *App) Listener(listener net.Listener) error {
	return http.Serve(listener, a.Handler())
}

func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := a.pool.Get().(*Ctx)
	ctx.reset()
	ctx.Request = req
	ctx.writer = newResponseWriter(w)
	ctx.Response = ctx.writer
	defer a.pool.Put(ctx)

	path := req.URL.Path
	if root := a.route.trees[req.Method]; root != nil {
		if value := root.getValue(path, a.route.getParams); value.handlers != nil {
			ctx.path = path
			ctx.handlers = value.handlers
			ctx.routePath = value.fullPath
			var err error
			if value.params != nil {
				ctx.Params = *value.params
				err = ctx.Next()
				a.route.putParams(value.params)
			} else {
				err = ctx.Next()
			}
			if err != nil {
				ctx.HandleError(err)
			}
			return
		}
	}
	ctx.handlers = a.route.allNoRouteHandlers
	serveError(ctx, http.StatusNotFound, default404Body)
}

func serveError(c *Ctx, code int, defaultMessage []byte) {
	c.writer.status = code
	_ = c.Next() //middlewares
	if c.Response.Committed() {
		return
	}
	c.Status(code)
	_, _ = c.Response.Write(defaultMessage)
}

func (app *App) NoRoute(handlers ...HandlerFunc) {
	app.route.noRoute(handlers...)
}
