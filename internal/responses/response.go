package response

import (
	"encoding/json"
	"net/http"
)

type ResponseBody struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err != nil {
		json.NewEncoder(w).Encode(ResponseBody{Data: data, Message: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(ResponseBody{Data: data, Message: "Success"})
}

// Error is a convenience function for sending an error response
func Error(w http.ResponseWriter, statusCode int, err error) {
	JSON(w, statusCode, nil, err)
}