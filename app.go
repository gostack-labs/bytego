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
	Router
}

func New() *App {
	return &App{
		Router: newRouter(),
	}
}

type HandlerFunc func(*Ctx)

func (a *App) Use(middlewares ...HandlerFunc) {
	a.Router.Use(middlewares...)
}

func (a *App) Handler() http.Handler {
	if !a.UseHTTP2 {
		return a
	}
	h2s := &http2.Server{}
	return h2c.NewHandler(a, h2s)
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
