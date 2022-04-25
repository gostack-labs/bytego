package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gostack-labs/bytego"
)

func New(config ...Config) bytego.HandlerFunc {
	cfg := DefaultConfig
	if len(config) > 0 { //config
		cfg = config[0]
		// Set default values
		if len(cfg.AllowMethods) == 0 {
			cfg.AllowMethods = DefaultConfig.AllowMethods
		}
		if len(cfg.AllowOrigins) == 0 {
			cfg.AllowOrigins = DefaultConfig.AllowOrigins
		}
		if cfg.MaxAge <= 0 {
			cfg.MaxAge = DefaultConfig.MaxAge
		}
	}
	allowMethods := strings.Join(cfg.AllowMethods, ",")
	allowHeaders := strings.Join(cfg.AllowHeaders, ",")
	exposeHeaders := strings.Join(cfg.ExposeHeaders, ",")

	maxAge := strconv.Itoa(cfg.MaxAge)

	return func(c *bytego.Ctx) error {
		origin := c.Header(bytego.HeaderOrigin)
		if len(origin) == 0 {
			return nil
		}
		allowOrigin := ""
		// Check allowed origins
		for _, o := range cfg.AllowOrigins {
			if o == "*" && cfg.AllowCredentials {
				allowOrigin = origin
				break
			}
			if o == "*" || o == origin {
				allowOrigin = o
				break
			}
			if matchSubdomain(origin, o) {
				allowOrigin = origin
				break
			}
		}

		if allowOrigin == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return nil
		}

		// Simple request
		if c.Request.Method != http.MethodOptions {
			c.AppendHeader(bytego.HeaderVary, bytego.HeaderOrigin)
			c.SetHeader(bytego.HeaderAccessControlAllowOrigin, allowOrigin)

			if cfg.AllowCredentials {
				c.SetHeader(bytego.HeaderAccessControlAllowCredentials, "true")
			}
			if exposeHeaders != "" {
				c.SetHeader(bytego.HeaderAccessControlExposeHeaders, exposeHeaders)
			}
			return c.Next()
		}

		// Options request
		// Preflight request
		c.AppendHeader(bytego.HeaderVary, bytego.HeaderOrigin)
		c.AppendHeader(bytego.HeaderVary, bytego.HeaderAccessControlRequestMethod)
		c.AppendHeader(bytego.HeaderVary, bytego.HeaderAccessControlRequestHeaders)
		c.SetHeader(bytego.HeaderAccessControlAllowOrigin, allowOrigin)
		c.SetHeader(bytego.HeaderAccessControlAllowMethods, allowMethods)

		// Set Allow-Credentials if set to true
		if cfg.AllowCredentials {
			c.SetHeader(bytego.HeaderAccessControlAllowCredentials, "true")
		}

		// Set Allow-Headers if not empty
		if allowHeaders != "" {
			c.SetHeader(bytego.HeaderAccessControlAllowHeaders, allowHeaders)
		} else {
			h := c.Header(bytego.HeaderAccessControlRequestHeaders)
			if h != "" {
				c.SetHeader(bytego.HeaderAccessControlAllowHeaders, h)
			}
		}

		// Set MaxAge
		if cfg.MaxAge > 0 {
			c.SetHeader(bytego.HeaderAccessControlMaxAge, maxAge)
		}
		c.AbortWithStatus(http.StatusNoContent)
		return nil
	}
}
