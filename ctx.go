package bytego

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type Ctx struct {
	path       string
	index      int
	handlers   []HandlerFunc
	Method     string
	Writer     http.ResponseWriter
	Request    *http.Request
	Params     Params
	sameSite   http.SameSite
	routerPath string
}

func (c *Ctx) reset() {
	c.index = -1
	c.handlers = nil
	c.Writer = nil
	c.Request = nil
}

func (c *Ctx) Param(key string) string {
	param, _ := c.Params.Get(key)
	return param
}

func (c *Ctx) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Ctx) Form(key string) string {
	c.Request.Cookies()
	return c.Request.FormValue(key)
}

func (c *Ctx) Status(code int) {
	c.Writer.WriteHeader(code)
}

func (c *Ctx) RouterPath() string {
	return c.routerPath
}

func (c *Ctx) String(code int, s string) {
	c.Status(code)
	_, _ = c.Writer.Write([]byte(s))
}

func (c *Ctx) JSON(code int, i interface{}) {
	// c.Status(code)
	if err := c.writeJSON(i); err != nil {
		panic(err)
	}
}

func (c *Ctx) JSONP(code int, i interface{}) {
	callback := c.Query("callback")
	if callback == "" {
		c.JSON(code, i)
		return
	}
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	_, err = c.Writer.Write(stringToBytes(callback))
	if err != nil {
		panic(err)
	}
	_, err = c.Writer.Write([]byte{'('})
	if err != nil {
		panic(err)
	}
	_, err = c.Writer.Write(b)
	if err != nil {
		panic(err)
	}
	_, err = c.Writer.Write([]byte{')', ';'})
	if err != nil {
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

func (c *Ctx) RemoteIP() string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

func (c *Ctx) ContentType() string {
	contentType := c.Request.Header.Get("Content-Type")
	for i, char := range contentType {
		if char == ' ' || char == ';' {
			return contentType[:i]
		}
	}
	return contentType
}

func (c *Ctx) SetSameSite(samesite http.SameSite) {
	c.sameSite = samesite
}

func (c *Ctx) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: c.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

func (c *Ctx) writeJSON(i interface{}) error {
	bs, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.writeContentType(c.Writer, "application/json; charset=utf-8")
	_, err = c.Writer.Write(bs)
	return err
}

func (c *Ctx) writeContentType(w http.ResponseWriter, contentType string) {
	header := w.Header()
	if headers := header["Content-Type"]; len(headers) == 0 {
		header["Content-Type"] = []string{contentType}
	}
}
