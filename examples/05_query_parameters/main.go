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
