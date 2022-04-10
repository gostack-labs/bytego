package recovery

import (
	"github.com/gostack-labs/bytego"
)

type RecoveryFunc func(c *bytego.Ctx, err interface{})

func Recover(fc RecoveryFunc) bytego.HandlerFunc {
	return func(c *bytego.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				c.Abort()
				fc(c, err)
			}
		}()
		return c.Next()
	}
}
