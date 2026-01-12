package main

import (
	"embed"
	"log"
	"log/slog"

	"github.com/cfichtmueller/srv"
	"github.com/cfichtmueller/srv/spa"
)

var (
	address = "127.0.0.1:8000"
	//go:embed static
	staticFS embed.FS
)

func main() {
	s := srv.NewServer()

	spah := spa.NewHandler(&spa.HandlerOpts{
		IndexFile:        "index.html",
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
