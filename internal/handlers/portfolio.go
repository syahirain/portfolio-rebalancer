package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"portfolio-rebalancer/internal/kafka"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/storage"
)

// HandlePortfolio handles new portfolio creation requests (feel free to update the request parameter/model)
// Sample Request (POST /portfolio):
//
//	{
//	    "user_id": "1",
//	    "allocation": {"stocks": 60, "bonds": 30, "gold": 10}
//	}
func HandlePortfolio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p models.Portfolio
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate UserID and Allocation
	if err := models.ValidateUserAndAllocation(p.UserID, p.Allocation); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := models.ValidatePercentage(p.Allocation); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save to Elasticsearch
	if err := storage.SavePortfolio(r.Context(), p); err != nil {
		log.Printf("Failed to save portfolio: %v", err)
		http.Error(w, "Failed to save portfolio", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// HandleRebalance handles portfolio rebalance requests from 3rd party provider (feel free to update the request parameter/model)
// Sample Request (POST /rebalance):
//
//	{
//	    "user_id": "1",
//	    "new_allocation": {"stocks": 70, "bonds": 20, "gold": 10}
//	}
func HandleRebalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UpdatedPortfolio
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate UserID and Allocation
	if err = models.ValidateUserAndAllocation(req.UserID, req.NewAllocation); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = models.ValidatePercentage(req.NewAllocation); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user exists in Elasticsearch
	if _, err = storage.GetPortfolio(r.Context(), req.UserID); err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("Failed to check user existence: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	log.Println("HandleRebalance==", req)

	// Marshal request to JSON for Kafka
	payload, err := json.Marshal(req)
	if err != nil {
		log.Printf("Failed to marshal request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Publish to Kafka
	if err := kafka.PublishMessage(r.Context(), payload); err != nil {
		log.Printf("Failed to publish message to Kafka: %v", err)
		http.Error(w, "Failed to queue rebalance request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
