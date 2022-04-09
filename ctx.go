package bytego

import (
	"encoding/json"
	"net/http"
)

type Ctx struct {
	path    string
	Method  string
	Writer  http.ResponseWriter
	Request *http.Request
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

func (c *Ctx) writeJSON(i interface{}) error {
	bs, err := json.Marshal(i)
	if err != nil {
		return err
	}
	_, err = c.Writer.Write(bs)
	return err
}
