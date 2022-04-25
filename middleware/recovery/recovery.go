package recovery

import (
	"fmt"
	"runtime"

	"github.com/gostack-labs/bytego"
)

type RecoveryFunc func(c *bytego.Ctx, err error)

func New() bytego.HandlerFunc {
	return Recover(func(c *bytego.Ctx, err error) {
		c.HandleError(err)
	})
}

func Recover(fc RecoveryFunc) bytego.HandlerFunc {
	return func(c *bytego.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				stack := make([]byte, 4<<10)
				length := runtime.Stack(stack, true)
				msg := fmt.Sprintf("[PANIC RECOVER] %v %s\n", err, stack[:length])
				c.Logger().Error(msg)
				c.Abort()
				fc(c, err)
			}
		}()
		return c.Next()
	}
}
