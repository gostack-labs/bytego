package bytego

import (
	"net/http"
)

type Group struct {
	basePath string
	route    Router
	handlers []HandlerFunc
}

func (g *Group) Group(relativePath string, handlers ...HandlerFunc) Router {
	return &Group{
		basePath: joinPath(g.basePath, relativePath),
		route:    g.route,
	}
}

func (g *Group) GET(relativePath string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodGet, relativePath, handlers...)
}

func (g *Group) POST(relativePath string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodPost, relativePath, handlers...)
}

func (g *Group) PUT(relativePath string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodPut, relativePath, handlers...)
}

func (g *Group) DELETE(relativePath string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodDelete, relativePath, handlers...)
}

func (g *Group) HEAD(relativePath string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodHead, relativePath, handlers...)
}

func (g *Group) PATCH(relativePath string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodPatch, relativePath, handlers...)
}

func (g *Group) OPTIONS(relativePath string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodOptions, relativePath, handlers...)
}
func (g *Group) TRACE(relativePath string, handlers ...HandlerFunc) Router {
	return g.Handle(http.MethodTrace, relativePath, handlers...)
}

func (g *Group) Any(relativePath string, handlers ...HandlerFunc) Router {
	for _, method := range anyMethods {
		g.Handle(method, relativePath, handlers...)
	}
	return g.route
}

func (g *Group) Use(middlewares ...HandlerFunc) {
	g.handlers = append(g.handlers, middlewares...)
}

func (g *Group) Handle(method string, relativePath string, handlers ...HandlerFunc) Router {
	path := joinPath(g.basePath, relativePath)
	return g.route.Handle(method, path, handlers...)
}

func (g *Group) Static(relativePath, root string) Router {
	path := joinPath(g.basePath, relativePath)
	return g.route.StaticFS(path, http.Dir(root))
}

func (g *Group) StaticFS(relativePath string, fsys http.FileSystem) Router {
	path := joinPath(g.basePath, relativePath)
	return g.route.StaticFS(path, fsys)
}

func (g *Group) StaticFile(relativePath, filepath string) Router {
	path := joinPath(g.basePath, relativePath)
	return g.route.StaticFile(path, filepath)
}
