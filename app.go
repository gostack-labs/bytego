package bytego

import (
	"context"
	"net"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type App struct {
	server   *http.Server
	UseHTTP2 bool
	route    Router
}

func New() *App {
	return &App{
		route: newRouter(),
	}
}

type HandlerFunc func(*Ctx)
type HandlersChain []HandlerFunc

func (a *App) GET(path string, handler HandlerFunc) Router {
	return a.route.GET(path, handler)
}

func (a *App) POST(path string, handler HandlerFunc) Router {
	return a.route.POST(path, handler)
}

func (a *App) PUT(path string, handler HandlerFunc) Router {
	return a.route.PUT(path, handler)
}

func (a *App) DELETE(path string, handler HandlerFunc) Router {
	return a.route.DELETE(path, handler)
}

func (a *App) HEAD(path string, handler HandlerFunc) Router {
	return a.route.HEAD(path, handler)
}
func (a *App) PATCH(path string, handler HandlerFunc) Router {
	return a.route.PATCH(path, handler)
}

func (a *App) OPTIONS(path string, handler HandlerFunc) Router {
	return a.route.OPTIONS(path, handler)
}

func (a *App) Any(path string, handler HandlerFunc) Router {
	return a.route.Any(path, handler)
}

func (a *App) Group(relativePath string, handlers ...HandlerFunc) *Group {
	return a.route.Group(relativePath, handlers...)
}

func (a *App) Handler() http.Handler {
	if !a.UseHTTP2 {
		return a.route
	}
	h2s := &http2.Server{}
	return h2c.NewHandler(a.route, h2s)
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

func (a *App) Listener(listener net.Listener) error {
	return http.Serve(listener, a.Handler())
}
