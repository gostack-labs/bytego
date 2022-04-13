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
	}
	allowMethods := strings.Join(cfg.AllowMethods, ",")
	allowHeaders := strings.Join(cfg.AllowHeaders, ",")
	exposeHeaders := strings.Join(cfg.ExposeHeaders, ",")

	maxAge := strconv.Itoa(cfg.MaxAge)

	return func(c *bytego.Ctx) error {
		origin := c.GetHeader(bytego.HeaderOrigin)
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
			c.Response.Header().Add(bytego.HeaderVary, bytego.HeaderOrigin)
			c.Header(bytego.HeaderAccessControlAllowOrigin, allowOrigin)

			if cfg.AllowCredentials {
				c.Header(bytego.HeaderAccessControlAllowCredentials, "true")
			}
			if exposeHeaders != "" {
				c.Header(bytego.HeaderAccessControlExposeHeaders, exposeHeaders)
			}
			return c.Next()
		}

		// Options request
		// Preflight request
		c.Response.Header().Add(bytego.HeaderVary, bytego.HeaderOrigin)
		c.Response.Header().Add(bytego.HeaderVary, bytego.HeaderAccessControlRequestMethod)
		c.Response.Header().Add(bytego.HeaderVary, bytego.HeaderAccessControlRequestHeaders)
		c.Header(bytego.HeaderAccessControlAllowOrigin, allowOrigin)
		c.Header(bytego.HeaderAccessControlAllowMethods, allowMethods)

		// Set Allow-Credentials if set to true
		if cfg.AllowCredentials {
			c.Header(bytego.HeaderAccessControlAllowCredentials, "true")
		}

		// Set Allow-Headers if not empty
		if allowHeaders != "" {
			c.Header(bytego.HeaderAccessControlAllowHeaders, allowHeaders)
		} else {
			h := c.GetHeader(bytego.HeaderAccessControlRequestHeaders)
			if h != "" {
				c.Header(bytego.HeaderAccessControlAllowHeaders, h)
			}
		}

		// Set MaxAge
		if cfg.MaxAge > 0 {
			c.Header(bytego.HeaderAccessControlMaxAge, maxAge)
		}
		c.Status(http.StatusNoContent)
		return nil
	}
}
