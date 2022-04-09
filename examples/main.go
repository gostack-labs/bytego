package main

import (
	"log"

	"github.com/gostack-labs/bytego"
)

func main() {
	app := bytego.New()
	app.GET("/", func(c *bytego.Ctx) {
		c.String(200, "hello, world!")
	})
	app.GET("/a", func(c *bytego.Ctx) {
		c.String(200, "this is a page!")
	})
	app.GET("/json", func(c *bytego.Ctx) {
		c.JSON(200, map[string]string{
			"a": "b",
			"c": "d",
		})
	})
	app.POST("/abc", func(c *bytego.Ctx) {
		c.String(200, "sss")
	})
	app.Any("/any", func(c *bytego.Ctx) {
		c.String(200, "this is any router")
	})
	g := app.Group("/group")
	g.GET("/hello", func(c *bytego.Ctx) {
		c.String(200, "hello, group!")
	})
	g.GET("/any", func(c *bytego.Ctx) {
		c.String(200, "hello, group any!")
	})

	if err := app.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
