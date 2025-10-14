# Using Groups

**Problem:** Group multiple endpoints beneath a common path

Groups allow you to organize related endpoints under a common URL prefix, making your API more organized and maintainable. This is particularly useful for RESTful APIs where you have resources with multiple operations.

## What This Example Demonstrates

- Creating route groups with a common prefix (`/users`)
- Defining multiple endpoints within a group
- Using path parameters to extract dynamic values from URLs
- Returning JSON responses with different data structures

## API Endpoints Created

- `GET /users` - Returns a list of all users
- `GET /users/{id}` - Returns a specific user by ID

## Code Explanation

```go
package main

import (
	"log"

	"github.com/cfichtmueller/srv"
)

func main() {
	s := srv.NewServer().Use(srv.LoggingMiddleware())

	usersGroup := s.Group("/users")

	usersGroup.GET("", func(c *srv.Context) *srv.Response {
		return srv.Respond().Json([]string{"John", "Jane", "Jim"})
	})

	usersGroup.GET("/{id}", func(c *srv.Context) *srv.Response {
		id := c.PathValue("id")
		return srv.Respond().Json(map[string]any{"id": id})
	})

	log.Fatal(s.ListenAndServe(":8080"))
}
```

## Testing the Endpoints

You can test these endpoints using curl or any HTTP client:

```bash
# Get all users
curl http://localhost:8080/users
# Response: ["John", "Jane", "Jim"]

# Get a specific user by ID
curl http://localhost:8080/users/123
# Response: {"id": "123"}
```

## Key Benefits of Using Groups

1. **Organization**: Related endpoints are grouped together logically
2. **Maintainability**: Easier to manage and modify related routes
3. **Middleware**: Groups can have their own middleware applied
4. **Code Reuse**: Common path prefixes are defined once
5. **Scalability**: Easy to add new endpoints to existing groups
