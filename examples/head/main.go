package main

import (
	"log"
	"os"

	"github.com/cfichtmueller/srv"
)

var users = []string{"Bob", "Alice", "Charlie"}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	s := srv.NewServer()

	s.GET("", func(c *srv.Context) *srv.Response {
		return srv.Respond().Json(map[string]any{
			"_links": map[string]any{
				"users": map[string]any{
					"href": "/users",
				},
			},
		})
	})

	s.HEAD("/users", func(c *srv.Context) *srv.Response {
		return srv.Respond().Json(users)
	})

	s.GET("/users", func(c *srv.Context) *srv.Response {
		return srv.Respond().Json(users)
	})

	if err := s.ListenAndServe("127.0.0.1:" + port); err != nil {
		log.Fatal(err)
	}
}
