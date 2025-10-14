# Basic server with health check

**Problem:** Expose `/healthz` returning JSON.  
**Pattern:** Handlers return `*srv.Response`; use `srv.Respond().Json(...)`.  
**Gotchas:** Don’t write to `http.ResponseWriter` directly.

```go
package main

import (
  "log"
  "github.com/cfichtmueller/srv"
)

func main() {
  s := srv.NewServer().Use(srv.LoggingMiddleware())
  s.GET("/healthz", func(c *srv.Context) *srv.Response {
    return srv.Respond().Json(map[string]any{"status":"ok"})
  })
  log.Fatal(s.ListenAndServe(":8080"))
}
```

**Expected**: `GET /healthz` -> `{"status": "ok"}` (200)
