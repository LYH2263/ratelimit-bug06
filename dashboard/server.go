package dashboard

import (
	"embed"
	"net/http"
)

//go:embed static/*
var staticFS embed.FS

type Server struct {
	Handler http.Handler
}

func New(h http.Handler) *Server {
	if h == nil {
		h = http.NotFoundHandler()
	}
	mux := http.NewServeMux()
	mux.Handle("/api/", h)
	mux.Handle("/", http.FileServer(http.FS(staticFS)))
	return &Server{Handler: mux}
}
