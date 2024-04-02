package routes

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func InitializeRouter() {
	router := mux.NewRouter()
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"*"})
	origins := handlers.AllowedOrigins([]string{"*"})

	baseRoute := router.PathPrefix("/").Subrouter()

	TokenRouter(baseRoute)

	ipPort := "0.0.0.0:9000"
	fmt.Print("Server running on " + ipPort + "\n")
	serverErr := http.ListenAndServe(ipPort, handlers.CORS(headers, methods, origins)(router))
	if serverErr != nil {
		log.Fatal(serverErr)
	}
}
