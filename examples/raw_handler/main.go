package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/cfichtmueller/srv"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	s := srv.NewServer()
	s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		path := r.URL.Path
		query := r.URL.Query()
		headers := r.Header
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res := fmt.Sprintf("Method: %s\nPath: %s\nQuery: %v\nHeaders: %v\nBody: %s", method, path, query, headers, string(body))
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(res))
	})

	if err := s.ListenAndServe("127.0.0.1:" + port); err != nil {
		log.Fatal(err)
	}
}
