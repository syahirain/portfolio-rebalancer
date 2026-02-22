package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/storage" // Added import
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

	// Validate UserID
	if p.UserID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Validate Allocation
	if len(p.Allocation) == 0 {
		http.Error(w, "Allocation is required", http.StatusBadRequest)
		return
	}

	var totalPercentage float64
	for _, percentage := range p.Allocation {
		if percentage < 0 {
			http.Error(w, "Percentage cannot be negative", http.StatusBadRequest)
			return
		}
		totalPercentage += percentage
	}

	if totalPercentage != 100 {
		http.Error(w, "Total allocation percentage must be 100%", http.StatusBadRequest)
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
	json.NewDecoder(r.Body).Decode(&req)

	log.Println("HandleRebalance==", req)

	// TODO: Add Logic here

	w.WriteHeader(http.StatusOK)
}