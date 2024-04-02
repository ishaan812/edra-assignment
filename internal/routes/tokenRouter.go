package routes

import (
	controller "go-server-template/internal/controllers"

	"github.com/gorilla/mux"
)

func TokenRouter(tokenSubrouter *mux.Router) {
	tokenSubrouter.HandleFunc("/keys", controller.GenerateKeyHandler).Methods("POST")
	tokenSubrouter.HandleFunc("/keys", controller.GetAvailableKeyHandler).Methods("GET")
	tokenSubrouter.HandleFunc("/keys/{id}", controller.GetKeyInfoHandler).Methods("HEAD")
	tokenSubrouter.HandleFunc("/keys/{id}", controller.DeleteKeyHandler).Methods("DELETE")
	tokenSubrouter.HandleFunc("/keys/{id}", controller.UnblockKeyHandler).Methods("PUT")
	tokenSubrouter.HandleFunc("/keepalive/{id}", controller.KeepAliveHandler).Methods("PUT")
}
