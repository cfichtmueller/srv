package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cfichtmueller/srv"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	s := srv.NewServer()
	s.ANY("/", func(c *srv.Context) *srv.Response {
		return respond(c, "/")
	})
	s.ANY("/test", func(c *srv.Context) *srv.Response {
		return respond(c, "/test")
	})
	s.ANY("/test/{id}", func(c *srv.Context) *srv.Response {
		id := c.PathValue("id")
		return respond(c, "/test/"+id)
	})

	if err := s.ListenAndServe("127.0.0.1:" + port); err != nil {
		log.Fatal(err)
	}
}

func respond(c *srv.Context, path string) *srv.Response {
	return srv.Respond().Text(fmt.Sprintf("Method: %s\nPath: %s", c.Request().Method, path))
}
