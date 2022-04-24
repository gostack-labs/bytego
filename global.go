package bytego

const (
	HeaderContentLength   = "Content-Length"
	HeaderOrigin          = "Origin"
	HeaderVary            = "Vary"
	HeaderXForwardedFor   = "X-Forwarded-For"
	HeaderXForwardedProto = "X-Forwarded-Proto"
	HeaderXRealIP         = "X-Real-Ip"
	HeaderXRequestID      = "X-Request-ID"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	jsonContentType = "application/json; charset=utf-8"
	xmlContentType  = "application/xml; charset=utf-8"
	htmlContentType = "text/html; charset=utf-8"
)

var (
	default404Body = []byte("404 page not found")
)
