# pprof

pprof middleware

## Example

```go
package main

import (
    "github.com/gostack-labs/bytego"
    "github.com/gostack-labs/bytego/middleware/pprof"
)
func main() {
    app := bytego.New()
    pprof.Register(app, pprof.WithPrefix("/debug/pprof"))
    _ = app.Run(":8080")
}
```

### pprof tools

use the pprof tool to look at the heap profile:

```bash
go tool pprof http://localhost:8080/debug/pprof/heap
```

Or to look at a 30-second CPU profile:

```bash
go tool pprof http://localhost:8080/debug/pprof/profile
```

Or to look at the goroutine blocking profile, after calling runtime.SetBlockProfileRate in your program:

```bash
go tool pprof http://localhost:8080/debug/pprof/block
```

Or to collect a 5-second execution trace:

```bash
wget http://localhost:8080/debug/pprof/trace?seconds=5
```
