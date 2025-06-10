# srv

A simple and flexible HTTP server framework for Go.

## Features

- Clean and intuitive API for defining routes and handlers
- Middleware support at both server and route group levels
- Route grouping for better organization
- Built on top of Go's standard `net/http` package

## Getting Started

Install the dependency.

```go
go get github.com/cfichtmueller/srv
```

Create your server.

```go
package main

import "github.com/cfichtmueller/srv"

func main() {
    s := srv.NewServer().Use(srv.LoggingMiddleware())

    s.GET("/api/motd", func(c *srv.Context) *srv.Response {
        return srv.Respond().Json(map[string]any {
            "message": "Hello World",
        })
    })

    if err := s.ListenAndServe(":8080"); err != nil {
        log.Fatal(err)
    }
}
```