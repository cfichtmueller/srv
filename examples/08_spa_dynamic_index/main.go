package main

import (
	"embed"
	"log"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/cfichtmueller/srv"
	"github.com/cfichtmueller/srv/spa"
)

var (
	address = "127.0.0.1:8009"
	//go:embed static
	staticFS embed.FS
)

func main() {
	s := srv.NewServer()

	spah := spa.NewHandler(&spa.HandlerOpts{
		IndexFile:           "index.html",
		IndexFileIsTemplate: true,
		IndexDataFn: func(r *http.Request) (any, error) {
			headers := make([]Header, 0)
			for k, v := range r.Header {
				headers = append(headers, Header{
					Name:  k,
					Value: v[0],
				})
			}
			slices.SortFunc(headers, func(a, b Header) int {
				return strings.Compare(a.Name, b.Name)
			})
			return IndexData{
				Time:    time.Now().Format("Mon Jan _2 15:04:05 2006"),
				Headers: headers,
			}, nil
		},
		IndexCacheHeader: "max-age=300",
		AssetFS:          staticFS,
		AssetFSPrefix:    "static/",
		AssetPrefixes:    []string{"/assets"},
		AssetCacheHeader: "max-age=300",
	})

	s.Handle("/", spah)

	slog.Info("Starting server", "addr", address)
	if err := s.ListenAndServe(address); err != nil {
		log.Fatal(err)
	}
}

type IndexData struct {
	Time    string
	Headers []Header
}

type Header struct {
	Name  string
	Value string
}
