package bytego

import (
	"net/http"
)

type Group struct {
	basePath string
	route    Router
}

func (g *Group) Group(relativePath string, handlers ...HandlerFunc) *Group {
	return &Group{
		basePath: joinPath(g.basePath, relativePath),
		route:    g.route,
	}
}

func (g *Group) GET(relativePath string, handler HandlerFunc) Router {
	return g.Handle(http.MethodGet, relativePath, handler)
}

func (g *Group) POST(relativePath string, handler HandlerFunc) Router {
	return g.Handle(http.MethodPost, relativePath, handler)
}

func (g *Group) PUT(relativePath string, handler HandlerFunc) Router {
	return g.Handle(http.MethodPut, relativePath, handler)
}

func (g *Group) DELETE(relativePath string, handler HandlerFunc) Router {
	return g.Handle(http.MethodDelete, relativePath, handler)
}

func (g *Group) HEAD(relativePath string, handler HandlerFunc) Router {
	return g.Handle(http.MethodHead, relativePath, handler)
}

func (g *Group) PATCH(relativePath string, handler HandlerFunc) Router {
	return g.Handle(http.MethodPatch, relativePath, handler)
}

func (g *Group) OPTIONS(relativePath string, handler HandlerFunc) Router {
	return g.Handle(http.MethodOptions, relativePath, handler)
}

func (g *Group) Any(relativePath string, handler HandlerFunc) Router {
	for _, method := range anyMethods {
		g.Handle(method, relativePath, handler)
	}
	return g.route
}

func (g *Group) Handle(method string, relativePath string, handler HandlerFunc) Router {
	path := joinPath(g.basePath, relativePath)
	return g.route.Handle(method, path, handler)
}
