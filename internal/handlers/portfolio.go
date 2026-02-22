package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"portfolio-rebalancer/internal/kafka"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/storage"
	"portfolio-rebalancer/internal/utils"
)

// HandlePortfolio handles new portfolio creation requests (feel free to update the request parameter/model)
// Sample Request (POST /portfolio):
//
//	{
//	    "user_id": "1",
//	    "allocation": {"stocks": 60, "bonds": 30, "gold": 10}
//	}
func HandlePortfolio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// Decode request body
	var p models.Portfolio
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate UserID and Allocation
	if err := models.ValidateUserAndAllocation(p.UserID, p.Allocation); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if err := models.ValidatePercentage(p.Allocation); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Save to Elasticsearch
	if err := storage.SavePortfolio(r.Context(), &p); err != nil {
		log.Printf("Failed to save portfolio: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to save portfolio",
		})
		return
	}

	// Success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    p,
		Message: "Portfolio request accepted",
	})
}

// HandleRebalance handles portfolio rebalance requests from 3rd party provider (feel free to update the request parameter/model)
// Sample Request (POST /rebalance):
//
//	{
//	    "user_id": "1",
//	    "new_allocation": {"stocks": 70, "bonds": 20, "gold": 10}
//	}
func HandleRebalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// Decode request body
	var req models.UpdatedPortfolio
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate UserID and Allocation
	if err := models.ValidateUserAndAllocation(req.UserID, req.NewAllocation); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if err := models.ValidatePercentage(req.NewAllocation); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Get current allocation from Elasticsearch
	p, err := storage.GetPortfolio(r.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.APIResponse{
				Success: false,
				Message: "User not found",
			})
			return
		}

		log.Printf("Failed to get current portfolio: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	log.Println("HandleRebalance==", req)

	// Check canonical hash
	newHash := utils.CanonicalHash(req.NewAllocation)
	currentHash := utils.CanonicalHash(p.Allocation)
	if newHash == currentHash {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "New allocation is the same as current allocation",
		})
		return
	}

	// Marshal request to JSON for Kafka
	rbk := models.RebalancePortfolioKafka{
		UserID:            req.UserID,
		NewAllocation:     req.NewAllocation,
		CurrentAllocation: p.Allocation,
	}

	payload, err := json.Marshal(rbk)
	if err != nil {
		log.Printf("Failed to marshal request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	// Publish to Kafka
	if err := kafka.PublishMessage(r.Context(), payload); err != nil {
		log.Printf("Failed to publish message to Kafka: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to queue rebalance request",
		})
		return
	}

	// Success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    req,
		Message: "Rebalance request accepted",
	})
}
