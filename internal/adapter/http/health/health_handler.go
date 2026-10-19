package http

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func NewHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode( HealthResponse{
			Status: "ok",
		}); err != nil {
			return 
		}
	}
}