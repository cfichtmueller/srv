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
