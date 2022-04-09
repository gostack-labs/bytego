package bytego

type Group struct {
	basePath string
	router   Router
}

func (g *Group) Group(relativePath string, handlers ...HandlerFunc) *Group {
	return &Group{
		basePath: g.basePath + relativePath,
		router:   g.router,
	}
}

func (g *Group) Get(relativePath string, handler HandlerFunc) Router {
	return g.router.Get(g.basePath+relativePath, handler)
}

func (g *Group) Post(relativePath string, handler HandlerFunc) Router {
	return g.router.Post(g.basePath+relativePath, handler)
}

func (g *Group) Put(relativePath string, handler HandlerFunc) Router {
	return g.router.Put(g.basePath+relativePath, handler)
}

func (g *Group) Delete(relativePath string, handler HandlerFunc) Router {
	return g.router.Delete(g.basePath+relativePath, handler)
}

func (g *Group) Head(relativePath string, handler HandlerFunc) Router {
	return g.router.Head(g.basePath+relativePath, handler)
}

func (g *Group) Options(relativePath string, handler HandlerFunc) Router {
	return g.router.Options(g.basePath+relativePath, handler)
}
