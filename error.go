package bytego

import "net/http"

type ErrorHandler func(error, *Ctx)
type ErrorCode interface {
	Error() string
	ErrCode() int
}

func defaultErrorHandler(err error, c *Ctx) {
	errCode, ok := err.(ErrorCode)
	var statusCode int
	var code int
	var msg string
	if ok { //normal
		statusCode = http.StatusOK
		code = errCode.ErrCode()
		msg = err.Error()
	} else {
		statusCode = http.StatusInternalServerError
		code = statusCode
		if c.isDebug {
			msg = err.Error()
		} else {
			msg = http.StatusText(http.StatusInternalServerError)
		}
	}

	if c.Request.Method == http.MethodHead {
		c.Status(statusCode)
	} else {
		_ = c.JSON(statusCode, Map{
			"code": code,
			"msg":  msg,
		})
	}
}
