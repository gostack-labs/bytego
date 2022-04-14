package pprof

import (
	"net/http"
	"net/http/pprof"

	"github.com/gostack-labs/bytego"
)

func Register(app *bytego.App, opts ...Options) {
	opt := &option{
		prefix: "/debug/pprof",
	}
	if len(opts) > 0 {
		for _, o := range opts {
			o(opt)
		}
	}
	//register
	g := app.Group(opt.prefix)
	{
		g.GET("/", pprofHandler(pprof.Index))
		g.GET("/cmdline", pprofHandler(pprof.Cmdline))
		g.GET("/profile", pprofHandler(pprof.Profile))
		g.POST("/symbol", pprofHandler(pprof.Symbol))
		g.GET("/symbol", pprofHandler(pprof.Symbol))
		g.GET("/trace", pprofHandler(pprof.Trace))
		g.GET("/allocs", pprofHandler(pprof.Handler("allocs").ServeHTTP))
		g.GET("/block", pprofHandler(pprof.Handler("block").ServeHTTP))
		g.GET("/goroutine", pprofHandler(pprof.Handler("goroutine").ServeHTTP))
		g.GET("/heap", pprofHandler(pprof.Handler("heap").ServeHTTP))
		g.GET("/mutex", pprofHandler(pprof.Handler("mutex").ServeHTTP))
		g.GET("/threadcreate", pprofHandler(pprof.Handler("threadcreate").ServeHTTP))
	}
}

type option struct {
	prefix string
}
type Options func(*option)

func WithPrefix(prefix string) Options {
	return func(o *option) {
		o.prefix = prefix
	}
}

func pprofHandler(h http.HandlerFunc) bytego.HandlerFunc {
	handler := http.HandlerFunc(h)
	return func(c *bytego.Ctx) error {
		handler.ServeHTTP(c.Response, c.Request)
		return nil
	}
}
