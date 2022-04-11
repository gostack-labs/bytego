package bytego

import (
	"encoding/json"
	"encoding/xml"
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
	isDebug    bool
	binder     *binder
}

const (
	jsonContentType = "application/json; charset=utf-8"
	xmlContentType  = "application/xml; charset=utf-8"
)

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

func (c *Ctx) String(code int, s string) error {
	c.Status(code)
	_, err := c.Writer.Write([]byte(s))
	return err
}

func (c *Ctx) JSON(code int, i interface{}) error {
	c.Status(code)
	bs, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.writeContentType(c.Writer, jsonContentType)
	_, err = c.Writer.Write(bs)
	return err
}

func (c *Ctx) JSONP(code int, i interface{}) error {
	callback := c.Query("callback")
	if callback == "" {
		return c.JSON(code, i)
	}
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.writeContentType(c.Writer, jsonContentType)
	if _, err = c.Writer.Write(stringToBytes(callback)); err != nil {
		return err
	}
	if _, err = c.Writer.Write([]byte{'('}); err != nil {
		return err
	}
	if _, err = c.Writer.Write(b); err != nil {
		return err
	}
	if _, err = c.Writer.Write([]byte{')', ';'}); err != nil {
		return err
	}
	return err
}

func (c *Ctx) XML(code int, i interface{}) error {
	bs, err := xml.Marshal(i)
	if err != nil {
		return err
	}
	c.writeContentType(c.Writer, xmlContentType)
	_, err = c.Writer.Write(bs)
	return err
}

func (c *Ctx) EmptyContent(code int) error {
	c.Status(code)
	return nil
}

func (c *Ctx) Next() error {
	c.index++
	for c.index < len(c.handlers) {
		err := c.handlers[c.index](c)
		c.index++
		if err != nil {
			return err
		}
	}
	return nil
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

func (c *Ctx) IsDebug() bool {
	return c.isDebug
}

func (c *Ctx) Bind(i interface{}) error {
	return c.binder.Bind(c, i)
}

func (c *Ctx) writeContentType(w http.ResponseWriter, contentType string) {
	header := w.Header()
	if headers := header["Content-Type"]; len(headers) == 0 {
		header["Content-Type"] = []string{contentType}
	}
}
