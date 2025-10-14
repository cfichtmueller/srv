package main

import (
	"log"

	"github.com/cfichtmueller/srv"
)

type Request struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (r *Request) Validate() error {
	v := srv.RequireNotEmpty("name", r.Name, nil)
	v = srv.RequireMinLength("name", 3, r.Name, v)
	v = srv.RequireMaxLength("name", 100, r.Name, v)
	v = srv.Require("age", "AgeTooLow", "Age must be greater than 0", r.Age > 0, v)
	v = srv.Require("age", "AgeTooHigh", "Age must be less than 100", r.Age < 100, v)
	return srv.Validate(v)
}

func main() {
	s := srv.NewServer().Use(srv.LoggingMiddleware())

	s.POST("/users", func(c *srv.Context) *srv.Response {
		var req Request
		// Bind the request body to req
		// If request implements Validate() error - the validation will be invoked
		// if BindJSON returns a response, return it
		// if BindJSON doesn't return a response, the binding was successful
		if res := c.BindJSON(&req); res != nil {
			return res
		}
		return srv.Respond().Json(req)
	})

	log.Fatal(s.ListenAndServe(":8080"))
}
