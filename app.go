package bytego

import (
	"context"
	"net"
	"net/http"
)

type App struct {
	server *http.Server
	route  *route
	Router
}

func New() *App {
	r := newRouter()
	return &App{
		route:  r,
		Router: r,
	}
}

type HandlerFunc func(*Ctx) error
type Map map[string]interface{}
type Validate func(i interface{}) error

func (a *App) Use(middlewares ...HandlerFunc) {
	a.Router.Use(middlewares...)
}

func (a *App) Handler() http.Handler {
	return a
}

func (a *App) Validator(fc Validate) {
	if fc == nil {
		return
	}
	a.route.binder.validate = fc
}

func (a *App) ErrorHandler(fc ErrorHandler) {
	if fc == nil {
		return
	}
	a.route.errorHandler = fc
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
	a.route.isDebug = isDebug
}

func (a *App) Listener(listener net.Listener) error {
	return http.Serve(listener, a.Handler())
}
