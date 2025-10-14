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
