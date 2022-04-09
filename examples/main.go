package main

import "github.com/gostack-labs/bytego"

func main() {
	app := bytego.New()
	app.Get("/", func(c *bytego.Ctx) {
		c.String(200, "hello, world!")
	})
	app.Get("/a", func(c *bytego.Ctx) {
		c.String(200, "this is a page!")
	})
	app.Get("/json", func(c *bytego.Ctx) {
		c.JSON(200, map[string]string{
			"a": "b",
			"c": "d",
		})
	})
	app.Post("/abc", func(c *bytego.Ctx) {
		c.String(200, "sss")
	})
	g := app.Group("/group")
	g.Get("/hello", func(c *bytego.Ctx) {
		c.String(200, "hello, group!")
	})
	app.Run(":8080")
}
