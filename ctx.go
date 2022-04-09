package bytego

import (
	"encoding/json"
	"net/http"
)

type Ctx struct {
	path     string
	index    int
	handlers []HandlerFunc
	Method   string
	Writer   http.ResponseWriter
	Request  *http.Request
}

func (c *Ctx) reset() {
	c.index = -1
	c.handlers = nil
	c.Writer = nil
	c.Request = nil
}

func (c *Ctx) Status(code int) {
	c.Writer.WriteHeader(code)
}
func (c *Ctx) String(code int, s string) {
	c.Status(code)
	_, _ = c.Writer.Write([]byte(s))
}

func (c *Ctx) JSON(code int, i interface{}) {
	c.Status(code)
	if err := c.writeJSON(i); err != nil {
		panic(err)
	}
}

func (c *Ctx) Next() {
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Ctx) Abort() {
	c.index = len(c.handlers) + 1
}
func (c *Ctx) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

func (c *Ctx) writeJSON(i interface{}) error {
	bs, err := json.Marshal(i)
	if err != nil {
		return err
	}
	_, err = c.Writer.Write(bs)
	return err
}
