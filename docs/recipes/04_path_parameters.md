# Path Parameters

**Problem**: How to extract parameters from a URL path.

Path parameters allow you to capture dynamic values from URL segments. In the `srv` framework, you define path parameters using curly braces `{paramName}` in your route patterns and extract them using `c.PathValue("paramName")`.

## Basic Usage

```go
package main

import (
	"log"

	"github.com/cfichtmueller/srv"
)

func main() {
	s := srv.NewServer().Use(srv.LoggingMiddleware())

	// Static route - no parameters
	s.GET("/users", func(c *srv.Context) *srv.Response {
		return srv.Respond().Json([]string{"John", "Jane", "Jim"})
	})

	// Dynamic route with single parameter
	s.GET("/users/{id}", func(c *srv.Context) *srv.Response {
		id := c.PathValue("id")
		return srv.Respond().Json(map[string]any{"id": id})
	})

	log.Fatal(s.ListenAndServe(":8080"))
}
```

## Multiple Path Parameters

You can define multiple parameters in a single route:

```go
// Route: /users/{userId}/posts/{postId}
s.GET("/users/{userId}/posts/{postId}", func(c *srv.Context) *srv.Response {
	userId := c.PathValue("userId")
	postId := c.PathValue("postId")
	
	return srv.Respond().Json(map[string]any{
		"userId": userId,
		"postId": postId,
	})
})
```

## Parameter Validation and Type Conversion

Path parameters are always returned as strings. You may need to validate or convert them:

```go
s.GET("/users/{id}", func(c *srv.Context) *srv.Response {
	idStr := c.PathValue("id")
	
	// Convert to integer (with error handling)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return srv.Respond().Status(400).Json(map[string]string{
			"error": "Invalid user ID format",
		})
	}
	
	return srv.Respond().Json(map[string]any{"id": id})
})
```

## Testing Path Parameters

Test your endpoints with different parameter values:

```bash
# Test static route
curl http://localhost:8080/users

# Test dynamic route
curl http://localhost:8080/users/123
curl http://localhost:8080/users/abc

# Test multiple parameters
curl http://localhost:8080/users/123/posts/456
```

## Key Points

- Path parameters are defined using `{paramName}` syntax in route patterns
- Extract parameters using `c.PathValue("paramName")`
- Parameters are always returned as strings
- Route matching is exact - `/users/{id}` matches `/users/123` but not `/users/123/extra`
- Parameter names must be unique within a route pattern