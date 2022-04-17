package bytego

import (
	"encoding/json"
	"encoding/xml"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Ctx struct {
	path         string
	index        int
	handlers     []HandlerFunc
	Method       string
	writer       *responseWriter
	Response     ResponseWriter
	Request      *http.Request
	Params       Params
	sameSite     http.SameSite
	routePath    string
	isDebug      bool
	binder       *binder
	errorHandler ErrorHandler
	errorHandled bool
}

func (c *Ctx) reset() {
	c.index = -1
	c.handlers = nil
	c.writer = nil
	c.Response = nil
	c.Request = nil
	c.errorHandled = false
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
	c.Status(code)
	bs, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.writeContentType(c.Response, jsonContentType)
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
	_, err = c.Response.Write(bs)
	return err
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
		c.errorHandler(err, c)
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

/************************************/
/***** GOLANG.ORG/X/NET/CONTEXT *****/
/************************************/

// Deadline always returns that there is no deadline (ok==false),
// maybe you want to use Request.Context().Deadline() instead.
func (c *Ctx) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done always returns nil (chan which will wait forever),
// if you want to abort your work when the connection was closed
// you should use Request.Context().Done() instead.
func (c *Ctx) Done() <-chan struct{} {
	return nil
}

// Err always returns nil, maybe you want to use Request.Context().Err() instead.
func (c *Ctx) Err() error {
	return nil
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *Ctx) Value(key interface{}) interface{} {
	if key == 0 {
		return c.Request
	}
	if keyAsString, ok := key.(string); ok {
		val, _ := c.Params.Get(keyAsString)
		return val
	}
	return nil
}
