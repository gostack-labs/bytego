package bytego

import (
	"context"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type App struct {
	server   *http.Server
	UseHTTP2 bool
	ctx      context.Context
	cancel   func()
	route    Router
}

func New() *App {
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		ctx:    ctx,
		cancel: cancel,
		route:  newRouter(),
	}
	return app
}

type HandlerFunc func(*Ctx)
type HandlersChain []HandlerFunc

func (a *App) Get(path string, handler HandlerFunc) Router {
	return a.route.Get(path, handler)
}

func (a *App) Post(path string, handler HandlerFunc) Router {
	return a.route.Post(path, handler)
}

func (a *App) Put(path string, handler HandlerFunc) Router {
	return a.route.Put(path, handler)
}

func (a *App) Delete(path string, handler HandlerFunc) Router {
	return a.route.Delete(path, handler)
}

func (a *App) Head(path string, handler HandlerFunc) Router {
	return a.route.Head(path, handler)
}

func (a *App) Options(path string, handler HandlerFunc) Router {
	return a.route.Options(path, handler)
}

func (a *App) Handler() http.Handler {
	if !a.UseHTTP2 {
		return a.route
	}
	h2s := &http2.Server{}
	return h2c.NewHandler(a.route, h2s)
}

func (a *App) Run(addr string) {
	a.server = &http.Server{
		Addr:    addr,
		Handler: a.Handler(),
	}
	go func() {
		if err := a.server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	<-a.ctx.Done()
	a.shutdown()
}

func (a *App) shutdown() {
	if err := a.server.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func (a *App) Stop() {
	a.cancel()
}

func (a *App) Listener(listener net.Listener) error {
	return http.Serve(listener, a.Handler())
}
