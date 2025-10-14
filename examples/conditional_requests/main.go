package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/cfichtmueller/srv"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ETag      string    `json:"-"`
}

var users = make([]*User, 0)

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
				if res := c.ConditionalIfNoneMatch(user.ETag); res != nil {
					return res
				}
				if res := c.ConditionalIfModifiedSince(user.CreatedAt); res != nil {
					return res
				}
				return srv.Respond().Json(user).LastModified(user.UpdatedAt).ETag(user.ETag)
			}
		}
		return srv.Respond().NotFound()
	})

	s.PUT("/users/{id}", func(c *srv.Context) *srv.Response {
		id := c.PathValue("id")
		for _, user := range users {
			if user.ID == id {
				if res := c.ConditionalIfMatch(user.ETag); res != nil {
					return res
				}
				return srv.Respond().Json(user).LastModified(user.UpdatedAt).ETag(user.ETag)
			}
		}
		return srv.Respond().NotFound()
	})

	if err := s.ListenAndServe("127.0.0.1:" + port); err != nil {
		log.Fatal(err)
	}
}

func init() {
	now := time.Now()
	users = append(users, &User{
		ID:        "john",
		Name:      "John Doe",
		CreatedAt: now.Add(-2 * time.Hour),
		UpdatedAt: now.Add(-1 * time.Hour),
		ETag:      strconv.FormatInt(now.UnixMilli(), 10),
	})

	users = append(users, &User{
		ID:        "jane",
		Name:      "Jane Doe",
		CreatedAt: now.Add(-4 * time.Hour),
		UpdatedAt: now.Add(-3 * time.Hour),
		ETag:      strconv.FormatInt(now.UnixMilli(), 10),
	})
}
