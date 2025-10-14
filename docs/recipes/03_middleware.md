# Middleware

This example demonstrates how to implement middleware in the `srv` web framework. Middleware functions allow you to intercept and modify HTTP requests and responses, enabling cross-cutting concerns like authentication, logging, and response modification.

**Key Concepts:**
- **Global Middleware**: Applied to all routes using `.Use()` method
- **Route-specific Middleware**: Applied to individual routes as additional parameters
- **Middleware Chain**: Middleware functions are executed in order, with each calling `next(c)` to continue the chain
- **Response Modification**: Middleware can modify responses or create their own

**Example Features:**
- Global logging middleware for all requests
- Custom header middleware that adds "X-Custom-Header" to responses
- Authentication middleware that checks for Authorization header


```go
package main

import (
	"log"

	"github.com/cfichtmueller/srv"
)

func main() {
	s := srv.NewServer().Use(srv.LoggingMiddleware())

	s.GET("/motd", func(c *srv.Context) *srv.Response {
		return srv.Respond().Json(map[string]any{"message": "Hello World"})
	}, customHeaderMiddleware, authMiddleware)

	log.Fatal(s.ListenAndServe(":8080"))
}

func customHeaderMiddleware(c *srv.Context, next srv.Handler) *srv.Response {
	return next(c).Header("X-Custom-Header", "Hello World")
}

func authMiddleware(c *srv.Context, next srv.Handler) *srv.Response {
	if c.Authorization() == "" {
		return srv.Respond().Unauthorized()
	}
	return next(c)
}

```
