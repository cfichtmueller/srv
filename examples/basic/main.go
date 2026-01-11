package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cfichtmueller/srv"
)

type CreateUserRequest struct {
	Name string `json:"name"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var users = make([]*User, 0)

func (r *CreateUserRequest) Validate() error {
	e := srv.RequireNotEmpty("name", r.Name, nil)
	return srv.Validate(e)
}

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
				"user": map[string]any{
					"href": "/users/{id}",
				},
			},
		})
	})

	s.GET("/users", func(c *srv.Context) *srv.Response {
		return srv.Respond().Json(users)
	})

	s.GET("/users/{id}", func(c *srv.Context) *srv.Response {
		id := c.PathValue("id")
		for _, user := range users {
			if user.ID == id {
				return srv.Respond().Json(user)
			}
		}
		return srv.Respond().NotFound()
	})

	s.POST("/users", func(c *srv.Context) *srv.Response {
		var req CreateUserRequest
		if res := c.BindJson(&req); res != nil {
			return res
		}
		id := fmt.Sprintf("user-%d", time.Now().UnixMilli())
		user := &User{
			ID:   id,
			Name: req.Name,
		}
		users = append(users, user)
		return srv.Respond().Json(user)
	})

	s.DELETE("/users/{id}", func(c *srv.Context) *srv.Response {
		id := c.PathValue("id")

		for i, user := range users {
			if user.ID == id {
				users = append(users[:i], users[i+1:]...)
			}
		}
		return srv.Respond().NoContent()
	})

	if err := s.ListenAndServe("127.0.0.1:" + port); err != nil {
		log.Fatal(err)
	}
}
