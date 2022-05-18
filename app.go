package bytego

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
)

type App struct {
	server *http.Server
	route  *router
	Router
	errorHandler ErrorHandler
	binder       *binder
	isDebug      bool
	render       Renderer
	Logger       Logger
}

func New() *App {
	r := newRouter()
	a := &App{
		route: r,
		Router: &Group{
			basePath: r.basePath,
			route:    r,
			isRoot:   true,
		},
		errorHandler: defaultErrorHandler,
		binder:       &binder{},
		Logger:       NewLogger(os.Stdout),
	}
	r.app = a
	return a
}

type HandlerFunc func(*Ctx) error
type Map map[string]interface{}
type Validate func(i interface{}) error
type ValidateTranslate func(c *Ctx, err error) error
type Renderer interface {
	Render(io.Writer, string, interface{}) error
}

func (a *App) Handler() http.Handler {
	return a.route
}

func (a *App) SetValidator(fc Validate, trans ...ValidateTranslate) {
	if fc == nil {
		return
	}
	a.binder.validate = fc
	if len(trans) > 0 {
		a.binder.validateTranslate = trans[0]
	}
}

func (a *App) SetErrorHandler(fc ErrorHandler) {
	if fc == nil {
		return
	}
	a.errorHandler = fc
}

func (a *App) SetRender(render Renderer) {
	if render == nil {
		return
	}
	a.render = render
}

func (a *App) SetLogger(l Logger) {
	if l == nil {
		return
	}
	a.Logger = l
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

func (app *App) NoRoute(handlers ...HandlerFunc) {
	app.route.noRoute(handlers...)
}
