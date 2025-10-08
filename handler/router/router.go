package router

import (
	"database/sql"
	"net/http"

	"github.com/TechBowl-japan/go-stations/handler"
	"github.com/TechBowl-japan/go-stations/handler/middleware"
	"github.com/TechBowl-japan/go-stations/service"
)

func NewRouter(todoDB *sql.DB) *http.ServeMux {
	// register routes
	mux := http.NewServeMux()

	mux.Handle("/healthz", middleware.UsarAgent(middleware.Authorization(&handler.HealthzHandler{})))

	// Initialize TODOService and pass it to TODOHandler
	todoService := service.NewTODOService(todoDB)
	mux.Handle("/todos", handler.NewTODOHandler(todoService))

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("panic test")
	})
	mux.Handle("/do-panic", middleware.UsarAgent(middleware.Recovery(panicHandler)))

	return mux
}
