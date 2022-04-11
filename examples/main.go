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
	// app.Validator(validator.New().Struct) //import github.com/go-playground/validator/v10
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
		return c.JSON(200, bytego.Map{
			"a": "b",
			"c": "d",
		})
	})
	type School struct {
		Name string `form:"schname"`
	}
	type City struct {
		CityName string
	}
	type WantJob struct {
		JobName string
	}
	type People struct {
		Name   string  `json:"name,omitempty"`
		Parent *People `json:"parent,omitempty"`
	}
	type Student struct {
		Name   string `xml:"name,omitempty" form:"formname"`
		Age    int    `xml:"age,omitempty" validate:"gte=0,lte=60"`
		School School `form:"sch"`
		City
		*WantJob
		Parent  *People
		Header1 string `header:"request-id"`
		Query1  string `query:"query1"`
		Param1  string `param:"id"`
	}
	app.GET("/xml", func(c *bytego.Ctx) error {
		return c.XML(200, &Student{Name: "hao", Age: 18})
	})
	app.GET("/jsonp", func(c *bytego.Ctx) error {
		return c.JSONP(200, bytego.Map{
			"a": "b",
			"c": "d",
		})
	})
	//curl -d '{"name":"a","age":22}' -H 'content-type:application/json' http://localhost:8080/bind/student
	//curl -d '<student><name>test</name><age>18</age></student>' -H 'content-type:application/xml' http://localhost:8080/bind/student
	//curl -d 'formname=test&age=18&sch.schname=aa' -H 'content-type:application/x-www-form-urlencoded' http://localhost:8080/bind/student
	//curl -d 'formname=test&age=18&sch.schname=aa'  http://localhost:8080/bind/student
	//curl -d 'formname=test&age=18&sch.schname=aa&cityname=hz&jobname=programer&parent.name=pname&parent.parent.name=ppname'  http://localhost:8080/bind/student
	app.POST("/bind/student/:id", func(c *bytego.Ctx) error {
		var s Student
		if err := c.Bind(&s); err != nil {
			return err
		}
		return c.JSON(200, s)
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
