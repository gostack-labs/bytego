package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gostack-labs/bytego"
	"github.com/gostack-labs/bytego/middleware/recovery"
)

type ErrorReult struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func (e ErrorReult) ErrCode() int {
	return e.Code
}
func (e ErrorReult) Error() string {
	return e.Msg
}

func NewErrorResult(code int, msg string, data ...interface{}) *ErrorReult {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	return &ErrorReult{
		Code: code,
		Msg:  msg,
		Data: d,
	}
}

func main() {
	app := bytego.New()
	app.Debug(true)
	app.Use(recovery.Recover(func(c *bytego.Ctx, err interface{}) {
		var errMsg string
		if e, ok := err.(error); ok {
			errMsg = e.Error()
		} else {
			errMsg = fmt.Sprintf("%v", err)
		}

		_ = c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code": http.StatusInternalServerError,
			"msg":  errMsg,
		})
	}))
	app.Use(func(c *bytego.Ctx) error {
		log.Println("log--pre")
		err := c.Next()
		log.Println("log--ok")
		return err
	})
	app.GET("/", func(c *bytego.Ctx) error {
		return c.String(200, "hello, world!")
	})
	app.GET("/a", func(c *bytego.Ctx) error {
		return c.String(200, "this is a page!")
	})
	app.GET("/error", func(c *bytego.Ctx) error {
		return errors.New("this is an error")
	})
	app.GET("/error2", func(c *bytego.Ctx) error {
		return NewErrorResult(10010, "这是一个错误码")
	})

	app.GET("/user", func(c *bytego.Ctx) error {
		return c.String(200, c.Query("id"))
	})
	app.GET("/user/:id", func(c *bytego.Ctx) error {
		return c.String(200, c.Param("id")+c.RouterPath())
	})
	app.POST("/user/update", func(c *bytego.Ctx) error {
		return c.String(200, c.Form("new_name"))
	})
	app.GET("/panic", func(c *bytego.Ctx) error {
		panic("this a error")
	})
	app.GET("/json", func(c *bytego.Ctx) error {
		return c.JSON(200, map[string]string{
			"a": "b",
			"c": "d",
		})
	})
	app.GET("/jsonp", func(c *bytego.Ctx) error {
		return c.JSONP(200, map[string]string{
			"a": "b",
			"c": "d",
		})
	})

	app.Any("/any", func(c *bytego.Ctx) error {
		return c.String(200, "this is any router")
	})
	g := app.Group("/group")
	middleware := func(c *bytego.Ctx) error {
		log.Println("middleware ---- call")
		var err error
		if true {
			// c.Next()
			// c.AbortWithStatus(500)
			// c.String(200, "abort")
			// c.Abort()
			err = c.Next()
		}
		log.Println("middleware ---- ok")
		return err
	}
	g.GET("/hello", middleware, func(c *bytego.Ctx) error {
		log.Println("router call")
		return c.String(200, "hello, group!")
	})
	g.GET("/any", func(c *bytego.Ctx) error {
		return c.String(200, "hello, group any!")
	})

	if err := app.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
