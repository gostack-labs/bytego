package bytego

import (
	"net/http"
	"path"
)

type Group struct {
	basePath string
	route    *router
	handlers []HandlerFunc
	isRoot   bool
}

func (g *Group) GET(path string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodGet, path, handlers...)
}

func (g *Group) POST(path string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodPost, path, handlers...)
}

func (g *Group) PUT(path string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodPut, path, handlers...)
}

func (g *Group) DELETE(path string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodDelete, path, handlers...)
}

func (g *Group) HEAD(path string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodHead, path, handlers...)
}

func (g *Group) PATCH(path string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodPatch, path, handlers...)
}

func (g *Group) OPTIONS(path string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodOptions, path, handlers...)
}

func (g *Group) TRACE(path string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodTrace, path, handlers...)
}

func (g *Group) Any(path string, handlers ...HandlerFunc) Router {
	for _, method := range anyMethods {
		g.Handle(method, path, handlers...)
	}
	return g.router()
}

func (g *Group) Handle(method string, path string, handlers ...HandlerFunc) Router {
	path = joinPath(g.basePath, path)
	handlers = combineHandlers(g.handlers, handlers)
	g.route.add(method, path, handlers...)
	return g.router()
}

func (g *Group) Group(relativePath string, handlers ...HandlerFunc) Router {
	return &Group{
		basePath: joinPath(g.basePath, relativePath),
		route:    g.route,
		handlers: combineHandlers(g.handlers, handlers),
	}
}

func (g *Group) Static(relativePath, root string) Router {
	return g.StaticFS(relativePath, http.Dir(root))
}

func (g *Group) StaticFS(relativePath string, fsys http.FileSystem) Router {
	prefix := joinPath(g.basePath, relativePath)
	fileServer := http.StripPrefix(prefix, http.FileServer(fsys))
	handlerFunc := func(c *Ctx) error {
		// filepath := c.Param("filepath")
		fileServer.ServeHTTP(c.Response, c.Request)
		return nil
	}
	return g.GET(path.Join(relativePath, "/*filepath"), handlerFunc)
}

func (g *Group) StaticFile(relativePath, filePath string) Router {
	handlerFunc := func(c *Ctx) error {
		http.ServeFile(c.Response, c.Request, filePath)
		return nil
	}
	return g.GET(relativePath, handlerFunc)
}

func (g *Group) Use(middlewares ...HandlerFunc) {
	g.handlers = append(g.handlers, middlewares...)
	g.route.rebuild404Handlers()
}

func (g *Group) router() Router {
	if g.isRoot {
		return g.route.app
	}
	return g
}
