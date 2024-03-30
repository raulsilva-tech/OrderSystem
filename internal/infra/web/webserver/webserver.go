package webserver

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type WebServer struct {
	Router        chi.Router
	Handlers      map[string]http.HandlerFunc
	WebServerPort string
	Server        *http.Server
}

func NewWebServer(serverPort string) *WebServer {
	router := chi.NewRouter()
	return &WebServer{
		Router:        router,
		Handlers:      make(map[string]http.HandlerFunc),
		WebServerPort: serverPort,
		Server:        &http.Server{Addr: serverPort, Handler: router},
	}
}

func (s *WebServer) AddHandler(path string, handler http.HandlerFunc) {
	s.Handlers[path] = handler
}

// loop through the handlers and add them to the router
// register middeleware logger
// start the server
func (s *WebServer) Start() {
	s.Router.Use(middleware.Logger)
	for path, handler := range s.Handlers {
		s.Router.Handle(path, handler)
	}
	if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Coult not listen to the server")
	}
}

func (s *WebServer) Shutdown(ctx context.Context) error {

	return s.Server.Shutdown(ctx)
}
