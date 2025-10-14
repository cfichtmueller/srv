# Query Parameters

**Problem**: How to extract and validate query parameters from HTTP requests.

Query parameters are key-value pairs in the URL after a `?` character (e.g., `?name=john&age=25`). The `srv` package provides several methods to safely extract and validate these parameters.

## Available Methods

### Basic Query Parameter Extraction

- `c.Query(key)` - Returns the raw string value of a query parameter
- `c.HasQuery(key)` - Returns true if the query parameter exists
- `c.StringQuery(key)` - Returns the URL-decoded string value with error handling
- `c.StringQueryOrDefault(key, defaultValue)` - Returns URL-decoded string or default value
- `c.IntQuery(key)` - Returns parsed integer value with error handling
- `c.IntQueryOrDefault(key, defaultValue)` - Returns parsed integer or default value

```go
package main

import (
	"log"

	"github.com/cfichtmueller/srv"
)

var users = []string{"John", "Jane", "Jim"}

func main() {
	s := srv.NewServer().Use(srv.LoggingMiddleware())

	s.GET("/users", func(c *srv.Context) *srv.Response {
		name := c.Query("name")
		if name == "" {
			limit, r := c.IntQueryOrDefault("limit", 10)
			if r != nil {
				return r
			}
			if limit > len(users) {
				limit = len(users)
			}
			return srv.Respond().Json(users[:limit])
		}
		return srv.Respond().Json([]string{name})
	})

	log.Fatal(s.ListenAndServe(":8080"))
}
```