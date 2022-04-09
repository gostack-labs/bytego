package main

import "github.com/gostack-labs/bytego"

func main() {
	app := bytego.New()
	app.Get("/", func(ctx *bytego.Ctx) {
		ctx.String(200, "hello, world!")
	})
	app.Get("/a", func(ctx *bytego.Ctx) {
		ctx.String(200, "this is a page!")
	})
	app.Get("/json", func(ctx *bytego.Ctx) {
		ctx.JSON(200, map[string]string{
			"a": "b",
			"c": "d",
		})
	})
	app.Post("/abc", func(ctx *bytego.Ctx) {
		ctx.String(200, "sss")
	})
	app.Run(":8080")
}
