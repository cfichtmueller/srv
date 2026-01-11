# LLM GUIDE

This file helps LLMs understand the `srv` web framework library.

## Mental Model

**Server**: The main HTTP server that handles requests and manages routes, middleware, and groups.

**Group**: Groups multiple endpoints under a common path prefix. Applies middleware to all subpaths. Useful for organizing RESTful APIs.

**Route**: A combination of an HTTP method (GET, POST, PUT, DELETE, etc.) and a path pattern.

**Middleware**: A function that wraps a handler or another middleware. Executes before the handler and can modify requests/responses.

**Handler**: A function that handles HTTP requests. Takes a `*srv.Context` and returns a `*srv.Response`.

**Context**: Contains request information, provides access to headers, query params, path params, and request body.

**Response**: Canonical response model with fluent API for setting status codes, headers, cookies, and body content.

## Basic Server Setup

```go
package main

import (
    "log"
    "github.com/cfichtmueller/srv"
)

func main() {
    s := srv.NewServer().Use(srv.LoggingMiddleware())
    s.GET("/healthz", func(c *srv.Context) *srv.Response {
        return srv.Respond().Json(map[string]any{"status": "ok"})
    })
    log.Fatal(s.ListenAndServe(":8080"))
}
```

## Handler Pattern

**CRITICAL**: Always return a `*srv.Response` via `srv.Respond()`. Never write to `http.ResponseWriter` directly.

```go
func handler(c *srv.Context) *srv.Response {
    return srv.Respond().Json(map[string]any{"message": "Hello World"})
}
```

## HTTP Methods

The server supports all standard HTTP methods:

```go
s.GET("/path", handler)
s.POST("/path", handler)
s.PUT("/path", handler)
s.DELETE("/path", handler)
s.PATCH("/path", handler)
s.HEAD("/path", handler)
s.OPTIONS("/path", handler)
s.ANY("/path", handler) // Matches any HTTP method
```

## Groups

Groups organize related endpoints under a common path prefix:

```go
usersGroup := s.Group("/users")
usersGroup.GET("", func(c *srv.Context) *srv.Response {
    return srv.Respond().Json([]string{"John", "Jane", "Jim"})
})
usersGroup.GET("/{id}", func(c *srv.Context) *srv.Response {
    id := c.PathValue("id")
    return srv.Respond().Json(map[string]any{"id": id})
})
```

Groups can be nested and have their own middleware:

```go
apiGroup := s.Group("/api", authMiddleware)
v1Group := apiGroup.Group("/v1", versionMiddleware)
```

## Path Parameters

Extract dynamic values from URL paths using `{paramName}` syntax:

```go
s.GET("/users/{id}", func(c *srv.Context) *srv.Response {
    id := c.PathValue("id") // Returns string
    return srv.Respond().Json(map[string]any{"id": id})
})

// Multiple parameters
s.GET("/users/{userId}/posts/{postId}", func(c *srv.Context) *srv.Response {
    userId := c.PathValue("userId")
    postId := c.PathValue("postId")
    return srv.Respond().Json(map[string]any{"userId": userId, "postId": postId})
})
```

**Note**: Path parameters are always strings. Convert to other types manually if needed.

## Query Parameters

Use helper methods to safely extract and validate query parameters:

```go
// Basic extraction
name := c.Query("name")                    // Raw string
hasName := c.HasQuery("name")             // Boolean

// Safe extraction with defaults
limit, err := c.IntQueryOrDefault("limit", 10)
if err != nil {
    return err // Returns *srv.Response with 400 Bad Request
}

name, err := c.StringQueryOrDefault("name", "default")
if err != nil {
    return err
}
```

## Request Body Binding

Bind JSON request bodies to structs:

```go
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func createUser(c *srv.Context) *srv.Response {
    var user User
    if res := c.BindJson(&user); res != nil {
        return res // Returns validation error response
    }
    return srv.Respond().Created(user)
}
```

## Validation

Implement the `Validatable` interface for automatic validation:

```go
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func (u *User) Validate() error {
    v := srv.RequireNotEmpty("name", u.Name, nil)
    v = srv.RequireMinLength("name", 3, u.Name, v)
    v = srv.RequireMaxLength("name", 100, u.Name, v)
    v = srv.Require("age", "AgeTooLow", "Age must be greater than 0", u.Age > 0, v)
    v = srv.Require("age", "AgeTooHigh", "Age must be less than 100", u.Age < 100, v)
    return srv.Validate(v)
}
```

### Validation Functions

- `RequireNotEmpty(field, value, prev)` - Check if string is not empty
- `RequireMinLength(field, min, value, prev)` - Check minimum string length
- `RequireMaxLength(field, max, value, prev)` - Check maximum string length
- `Require(field, code, message, condition, prev)` - Custom validation condition
- `RequireEnumValue(field, value, allowed, prev)` - Check if value is in allowed list
- `RequireRegex(field, value, pattern, prev)` - Check regex pattern match
- `RequireNotEmptySlice(field, slice, prev)` - Check if slice is not empty
- `RequireMinLengthSlice(field, min, slice, prev)` - Check minimum slice length
- `RequireMaxLengthSlice(field, max, slice, prev)` - Check maximum slice length

## Response Building

Use the fluent API to build responses:

```go
// Status codes
return srv.Respond().Status(200)
return srv.Respond().Created(data)
return srv.Respond().BadRequest(error)
return srv.Respond().NotFound()
return srv.Respond().Unauthorized()
return srv.Respond().Forbidden()
return srv.Respond().InternalServerError()

// Response body
return srv.Respond().Json(data)
return srv.Respond().Html("<h1>Hello</h1>")
return srv.Respond().Text("plain text")
return srv.Respond().Body("application/octet-stream", bytes)

// Headers
return srv.Respond().Header("X-Custom", "value")
return srv.Respond().ContentType("application/json")
return srv.Respond().CacheControl("no-cache")
return srv.Respond().ETag("abc123")
return srv.Respond().LastModified(time.Now())

// Cookies
return srv.Respond().Cookie("session", "value", 3600, "/", "", true, true)

// CORS
return srv.Respond().AccessControlAllowOrigin("*")
return srv.Respond().AccessControlAllowMethods("GET", "POST")
return srv.Respond().AccessControlAllowHeaders("Content-Type", "Authorization")
```

## Middleware

Middleware functions wrap handlers and can modify requests/responses:

```go
func authMiddleware(c *srv.Context, next srv.Handler) *srv.Response {
    if c.Authorization() == "" {
        return srv.Respond().Unauthorized()
    }
    return next(c)
}

func customHeaderMiddleware(c *srv.Context, next srv.Handler) *srv.Response {
    return next(c).Header("X-Custom-Header", "Hello World")
}

// Apply middleware
s.Use(srv.LoggingMiddleware()) // Global middleware
s.GET("/path", handler, authMiddleware) // Route-specific middleware
```

## Conditional Requests (Caching)

Use conditional methods to implement HTTP caching:

```go
func getResource(c *srv.Context) *srv.Response {
    etag := "abc123"
    lastModified := time.Now().Add(-time.Hour)
    
    // Check If-None-Match header
    if res := c.ConditionalIfNoneMatch(etag); res != nil {
        return res // Returns 304 Not Modified
    }
    
    // Check If-Modified-Since header
    if res := c.ConditionalIfModifiedSince(lastModified); res != nil {
        return res // Returns 304 Not Modified
    }
    
    return srv.Respond().Json(data).ETag(etag).LastModified(lastModified)
}
```

## Request Information

Access request data through the context:

```go
// Headers
auth := c.Authorization()
userAgent := c.UserAgent()
contentType := c.ContentType()
accept := c.Accept()

// Client information
clientIP := c.ClientIP()    // Resolves through proxy headers
remoteIP := c.RemoteIP()    // Direct remote address

// Cookies
session, err := c.Cookie("session")

// Form data
formValues := c.FormValues()

// Raw request body
data, err := c.GetRawData()
```

## Server Configuration

Configure the server:

```go
s := srv.NewServer()
    .SetMaxMultipartMemory(32 << 20) // 32MB
    .SetRemoteIPHeaders("X-Forwarded-For", "X-Real-IP")
    .SetTrustRemoteIdHeaders(true)
    .Use(srv.LoggingMiddleware())
```

## Error Handling

The framework provides structured error responses:

```go
type ErrorDto struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// Validation errors include field-level details
type ValidationError struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Errors  []Violation `json:"errors"`
}

type Violation struct {
    Field   string `json:"field"`
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

## Best Practices

1. **Always return `*srv.Response`** - Never write to `http.ResponseWriter` directly
2. **Use validation** - Implement `Validatable` interface for automatic request validation
3. **Handle errors gracefully** - Check return values from query parameter methods
4. **Use groups** - Organize related endpoints under common path prefixes
5. **Apply middleware** - Use middleware for cross-cutting concerns like auth, logging
6. **Implement caching** - Use conditional request methods for HTTP caching
7. **Type safety** - Convert path parameters to appropriate types when needed
8. **Error responses** - Use appropriate HTTP status codes and structured error responses