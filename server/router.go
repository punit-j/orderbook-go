package server

import (
	"net/http"
	"orders-manager/service"

	"github.com/gorilla/handlers"

	"github.com/gorilla/mux"
)

// InitRouter initializes the router for the server
func InitRouter(deps *Dependencies) (router *mux.Router, handler http.Handler) {
	router = mux.NewRouter()

	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	handler = handlers.CORS(headers, methods, origins)(router)

	// Add routes here
	router.HandleFunc(("/orders"), service.SubmitOrder(deps.orderService)).Methods(http.MethodPost)
	return router, handler
}
