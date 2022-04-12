package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gostack-labs/bytego"
	"github.com/gostack-labs/bytego/internal/fasttemplate"
)

func New(config ...Config) bytego.HandlerFunc {
	cfg := configDefault(config...)

	// Get timezone location
	if tz, err := time.LoadLocation(cfg.TimeZone); err != nil || tz == nil {
		cfg.timeZoneLocation = time.Local
	} else {
		cfg.timeZoneLocation = tz
	}
	template := fasttemplate.New(cfg.Format, "${", "}")
	pool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}

	return func(c *bytego.Ctx) error {
		start := time.Now()
		err := c.Next()
		if err != nil {
			c.HandleError(err)
		}
		stop := time.Now()
		pid := os.Getegid()
		buf := pool.Get().(*bytes.Buffer)
		buf.Reset()
		defer pool.Put(buf)
		_, _ = template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
			switch tag {
			case "time_unix":
				return buf.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
			case "time_unix_nano":
				return buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
			case "time_rfc3339":
				return buf.WriteString(time.Now().Format(time.RFC3339))
			case "time_rfc3339_nano":
				return buf.WriteString(time.Now().Format(time.RFC3339Nano))
			case "time_custom":
				return buf.WriteString(time.Now().In(cfg.timeZoneLocation).Format(cfg.TimeFormat))
			case "pid":
				return buf.WriteString(strconv.Itoa(pid))
			case "request_id":
				id := c.GetHeader(bytego.HeaderXRequestID)
				if id == "" {
					id = c.Response.Header().Get(bytego.HeaderXRequestID)
				}
				return buf.WriteString(id)
			case "remote_ip":
				return buf.WriteString(c.ClientIP())
			case "host":
				return buf.WriteString(c.Request.Host)
			case "uri":
				return buf.WriteString(c.Request.RequestURI)
			case "method":
				return buf.WriteString(c.Request.Method)
			case "path":
				p := c.Request.URL.Path
				if p == "" {
					p = "/"
				}
				return buf.WriteString(p)
			case "protocol":
				return buf.WriteString(c.Request.Proto)
			case "referer":
				return buf.WriteString(c.Request.Referer())
			case "user_agent":
				return buf.WriteString(c.Request.UserAgent())
			case "status":
				return buf.WriteString(strconv.Itoa(c.Response.Status()))
			case "error":
				if err != nil {
					// Error may contain invalid JSON e.g. `"`
					b, _ := json.Marshal(err.Error())
					b = b[1 : len(b)-1]
					return buf.Write(b)
				}
			case "latency":
				l := stop.Sub(start)
				return buf.WriteString(strconv.FormatInt(int64(l), 10))
			case "latency_human":
				return buf.WriteString(stop.Sub(start).String())
			case "bytes_in":
				cl := c.GetHeader(bytego.HeaderContentLength)
				if cl == "" {
					cl = "0"
				}
				return buf.WriteString(cl)
			case "bytes_out":
				return buf.WriteString(strconv.Itoa(c.Response.Size()))
			default:
				switch {
				case strings.HasPrefix(tag, "header:"):
					return buf.Write([]byte(c.GetHeader(tag[7:])))
				case strings.HasPrefix(tag, "query:"):
					return buf.Write([]byte(c.Query(tag[6:])))
				case strings.HasPrefix(tag, "form:"):
					return buf.Write([]byte(c.Form(tag[5:])))
				case strings.HasPrefix(tag, "cookie:"):
					cookie, err := c.Cookie(tag[7:])
					if err == nil {
						return buf.Write([]byte(cookie.Value))
					}
				}
			}
			return 0, nil
		})
		_, err = buf.WriteTo(cfg.Output)
		//cfg.Output.Write(buf.Bytes())
		return nil
	}
}
