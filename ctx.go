package bytego

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Ctx struct {
	app          *App
	path         string
	index        int
	handlers     []HandlerFunc
	writer       *responseWriter
	Response     ResponseWriter
	Request      *http.Request
	Params       Params
	sameSite     http.SameSite
	routePath    string
	errorHandled bool
	m            Map
	mu           sync.RWMutex
}

func (c *Ctx) reset() {
	c.index = -1
	c.handlers = nil
	c.Params = c.Params[:0]
	c.writer = nil
	c.Response = nil
	c.Request = nil
	c.errorHandled = false
	c.routePath = ""
	c.m = nil
}

func (c *Ctx) Param(key string) string {
	param, _ := c.Params.Get(key)
	return param
}

func (c *Ctx) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Ctx) Form(key string) string {
	return c.Request.FormValue(key)
}

func (c *Ctx) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

func (c *Ctx) Status(code int) {
	c.Response.WriteHeader(code)
}

func (c *Ctx) RoutePath() string {
	return c.routePath
}

func (c *Ctx) Header(key, value string) {
	if value == "" {
		c.Response.Header().Del(key)
		return
	}
	c.Response.Header().Set(key, value)
}

func (c *Ctx) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Ctx) AppendHeader(key string, values ...string) {
	if len(values) == 0 {
		return
	}
	h := c.Response.Header().Get(key)
	originalH := h
	for _, value := range values {
		if len(h) == 0 {
			h = value
		} else if h != value && !strings.HasPrefix(h, value+",") && !strings.HasSuffix(h, " "+value) &&
			!strings.Contains(h, " "+value+",") {
			h += ", " + value
		}
	}
	if originalH != h {
		c.Header(key, h)
	}
}

func (c *Ctx) String(code int, s string) error {
	c.Status(code)
	_, err := c.Response.Write([]byte(s))
	return err
}

func (c *Ctx) JSON(code int, i interface{}) error {
	bs, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.writeContentType(c.Response, jsonContentType)
	c.Status(code)
	_, err = c.Response.Write(bs)
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
	c.writeContentType(c.Response, jsonContentType)
	c.Status(code)
	if _, err = c.Response.Write(stringToBytes(callback)); err != nil {
		return err
	}
	if _, err = c.Response.Write([]byte{'('}); err != nil {
		return err
	}
	if _, err = c.Response.Write(b); err != nil {
		return err
	}
	if _, err = c.Response.Write([]byte{')', ';'}); err != nil {
		return err
	}
	return err
}

func (c *Ctx) XML(code int, i interface{}) error {
	bs, err := xml.Marshal(i)
	if err != nil {
		return err
	}
	c.writeContentType(c.Response, xmlContentType)
	c.Status(code)
	_, err = c.Response.Write(bs)
	return err
}

func (c *Ctx) HTML(code int, html string) error {
	return c.HTMLBlob(code, []byte(html))
}

func (c *Ctx) HTMLBlob(code int, b []byte) (err error) {
	c.writeContentType(c.Response, htmlContentType)
	c.Status(code)
	_, err = c.Response.Write(b)
	return err
}

func (c *Ctx) View(code int, name string, data interface{}) error {
	buf := new(bytes.Buffer)
	if c.app.render == nil {
		return errors.New("error: render template not set")
	}
	if err := c.app.render.Render(buf, name, data); err != nil {
		return err
	}
	return c.HTMLBlob(code, buf.Bytes())
}

func (c *Ctx) Redirect(location string, code ...int) error {
	status := http.StatusFound
	if len(code) > 0 {
		status = code[0]
		if status < http.StatusMultipleChoices || status > http.StatusPermanentRedirect {
			return fmt.Errorf("error redirect with status code %d", status)
		}
	}
	http.Redirect(c.Response, c.Request, location, status)
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

func (c *Ctx) HandleError(err error) {
	if !c.errorHandled {
		c.app.errorHandler(err, c)
	}
	c.errorHandled = true
}

func (c *Ctx) RemoteIP() string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

func (c *Ctx) ClientIP() string {
	if ip := c.Request.Header.Get(HeaderXForwardedFor); ip != "" {
		i := strings.IndexAny(ip, ",")
		if i > 0 {
			return strings.TrimSpace(ip[:i])
		}
		return ip
	}
	if ip := c.Request.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	return c.RemoteIP()
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
	http.SetCookie(c.Response, &http.Cookie{
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
	return c.app.isDebug
}

func (c *Ctx) Bind(i interface{}) error {
	return c.app.binder.Bind(c, i)
}

func (c *Ctx) writeContentType(w http.ResponseWriter, contentType string) {
	header := w.Header()
	if headers := header["Content-Type"]; len(headers) == 0 {
		header["Content-Type"] = []string{contentType}
	}
}

func (c *Ctx) Context() context.Context {
	if c.Request != nil {
		return c.Request.Context()
	}
	return context.Background()
}

func (c *Ctx) Set(key string, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = make(Map)
	}
	c.m[key] = val
}

func (c *Ctx) Get(key string) (val interface{}, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, exists = c.m[key]
	return
}
