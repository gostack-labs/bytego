package cors

import (
	"net/http"
)

type Config struct {
	// AllowOrigin defines a list of origins that may access the resource.
	// Default value []string{"*"}
	AllowOrigins []string

	// request `Access-Control-Allow-Methods` header value
	// Default value []string{"GET","POST","HEAD","PUT","DELETE","PATCH"}
	AllowMethods []string

	// AllowHeaders defines a list of request headers that can be used when
	// making the actual request. This is in response to a preflight request.
	//
	// Optional. Default value "".
	AllowHeaders []string
	// AllowCredentials indicates whether or not the response to the request
	// can be exposed when the credentials flag is true. When used as part of
	// a response to a preflight request, this indicates whether or not the
	// actual request can be made using credentials.
	//
	// Optional. Default value false.
	AllowCredentials bool

	// ExposeHeaders defines a whitelist headers that clients are allowed to
	// access.
	//
	// Optional. Default value "".
	ExposeHeaders []string

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached.
	//
	// Optional. Default value 0.
	MaxAge int
}

var DefaultConfig = Config{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodHead,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	},
	AllowHeaders:     []string{"Content-Type"},
	AllowCredentials: false,
	ExposeHeaders:    []string{},
	MaxAge:           12 * 60 * 60,
}
