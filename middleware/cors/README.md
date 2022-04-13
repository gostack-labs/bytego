# CORS Middleware

[Cross-Origin Resource Sharing](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)

## Examples

```go
package main
import (
  "github.com/gostack-labs/bytego"
  "github.com/gostack-labs/bytego/middleware/cors"
)

func main() {
    app := bytego.New()

    //Default config
    app.Use(cors.New())

    //Custom config
    app.Use(cors.New(cors.Config{
        AllowOrigins: []string{"https://bytego.dev", "https://github.com"},
        AllowHeaders:  []string{"Origin", "Content-Type", "Accept"},
    }))
    app.GET("/", func(c *bytego.Ctx) error {
        return c.String(200, "hello")
    })
    _ = app.Run(":8080")
}
```
