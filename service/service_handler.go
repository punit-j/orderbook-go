package service

import (
	"encoding/json"
	"net/http"
	"orders-manager/models"
	"time"
)

func SubmitOrder(service OrderService) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var order models.Order
		err := json.NewDecoder(req.Body).Decode(&order)
		if err != nil {
			http.Error(rw, "Invalid request payload", http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		order.Timestamp = time.Now().UTC()
		order.CreatedAt = time.Now().Unix()

		addedOrder, err := service.AddOrder(&order)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)
		json.NewEncoder(rw).Encode(addedOrder)
	}
}
