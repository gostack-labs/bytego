package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gostack-labs/bytego"
	"github.com/gostack-labs/bytego/middleware"
)

func main() {
	app := bytego.New()
	app.Use(middleware.Recover(func(c *bytego.Ctx, err interface{}) {
		var errMsg string
		if e, ok := err.(error); ok {
			errMsg = e.Error()
		} else {
			errMsg = fmt.Sprintf("%v", err)
		}

		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code": http.StatusInternalServerError,
			"msg":  errMsg,
		})
	}))
	app.Use(func(c *bytego.Ctx) {
		log.Println("log--pre")
		c.Next()
		log.Println("log--ok")
	})
	app.GET("/", func(c *bytego.Ctx) {
		c.String(200, "hello, world!")
	})
	app.GET("/a", func(c *bytego.Ctx) {
		c.String(200, "this is a page!")
	})
	app.GET("/error", func(c *bytego.Ctx) {
		panic("this a error")
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
	middleware := func(c *bytego.Ctx) {
		log.Println("middleware ---- call")
		if true {
			// c.Next()
			// c.AbortWithStatus(500)
			// c.String(200, "abort")
			// c.Abort()
			c.Next()
		}
		log.Println("middleware ---- ok")
	}
	g.GET("/hello", middleware, func(c *bytego.Ctx) {
		log.Println("router call")
		c.String(200, "hello, group!")
	})
	g.GET("/any", func(c *bytego.Ctx) {
		c.String(200, "hello, group any!")
	})

	if err := app.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
