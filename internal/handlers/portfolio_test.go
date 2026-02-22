package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/storage"
)

func TestHandlePortfolio(t *testing.T) {
	// Backup original function and restore after test
	origSave := savePortfolio
	defer func() { savePortfolio = origSave }()

	tests := []struct {
		name           string
		method         string
		body           interface{}
		mockSave       func(ctx context.Context, p *models.Portfolio) error
		expectedStatus int
	}{
		{
			name:   "Success",
			method: http.MethodPost,
			body: models.Portfolio{
				UserID:     "user1",
				Allocation: map[string]float64{"stocks": 60, "bonds": 40},
			},
			mockSave: func(ctx context.Context, p *models.Portfolio) error {
				return nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Invalid Method",
			method: http.MethodGet,
			body:   nil,
			mockSave: func(ctx context.Context, p *models.Portfolio) error {
				return nil
			},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid Body",
			method:         http.MethodPost,
			body:           "invalid-json",
			mockSave:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Validation Error - Missing UserID",
			method: http.MethodPost,
			body: models.Portfolio{
				UserID:     "",
				Allocation: map[string]float64{"stocks": 60, "bonds": 40},
			},
			mockSave:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Validation Error - Invalid Percentage",
			method: http.MethodPost,
			body: models.Portfolio{
				UserID:     "user1",
				Allocation: map[string]float64{"stocks": 60, "bonds": 30}, // 90%
			},
			mockSave:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Storage Error",
			method: http.MethodPost,
			body: models.Portfolio{
				UserID:     "user1",
				Allocation: map[string]float64{"stocks": 60, "bonds": 40},
			},
			mockSave: func(ctx context.Context, p *models.Portfolio) error {
				return errors.New("db error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set mock
			savePortfolio = tt.mockSave

			var reqBody []byte
			var err error
			if s, ok := tt.body.(string); ok {
				reqBody = []byte(s)
			} else if tt.body != nil {
				reqBody, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/portfolio", bytes.NewReader(reqBody))
			w := httptest.NewRecorder()

			HandlePortfolio(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleRebalance(t *testing.T) {
	// Backup original functions
	origGet := getPortfolio
	origPublish := publishMessage
	defer func() {
		getPortfolio = origGet
		publishMessage = origPublish
	}()

	tests := []struct {
		name           string
		method         string
		body           interface{}
		mockGet        func(ctx context.Context, userID string) (*models.Portfolio, error)
		mockPublish    func(ctx context.Context, payload []byte) error
		expectedStatus int
	}{
		{
			name:   "Success",
			method: http.MethodPost,
			body: models.UpdatedPortfolio{
				UserID:        "user1",
				NewAllocation: map[string]float64{"stocks": 70, "bonds": 30},
			},
			mockGet: func(ctx context.Context, userID string) (*models.Portfolio, error) {
				return &models.Portfolio{
					UserID:     "user1",
					Allocation: map[string]float64{"stocks": 60, "bonds": 40},
				}, nil
			},
			mockPublish: func(ctx context.Context, payload []byte) error {
				return nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "User Not Found",
			method: http.MethodPost,
			body: models.UpdatedPortfolio{
				UserID:        "user1",
				NewAllocation: map[string]float64{"stocks": 70, "bonds": 30},
			},
			mockGet: func(ctx context.Context, userID string) (*models.Portfolio, error) {
				return nil, storage.ErrUserNotFound
			},
			mockPublish:    nil,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Same Allocation",
			method: http.MethodPost,
			body: models.UpdatedPortfolio{
				UserID:        "user1",
				NewAllocation: map[string]float64{"stocks": 60, "bonds": 40},
			},
			mockGet: func(ctx context.Context, userID string) (*models.Portfolio, error) {
				return &models.Portfolio{
					UserID:     "user1",
					Allocation: map[string]float64{"stocks": 60, "bonds": 40},
				}, nil
			},
			mockPublish:    nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Kafka Publish Error",
			method: http.MethodPost,
			body: models.UpdatedPortfolio{
				UserID:        "user1",
				NewAllocation: map[string]float64{"stocks": 70, "bonds": 30},
			},
			mockGet: func(ctx context.Context, userID string) (*models.Portfolio, error) {
				return &models.Portfolio{
					UserID:     "user1",
					Allocation: map[string]float64{"stocks": 60, "bonds": 40},
				}, nil
			},
			mockPublish: func(ctx context.Context, payload []byte) error {
				return errors.New("kafka error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getPortfolio = tt.mockGet
			publishMessage = tt.mockPublish

			var reqBody []byte
			var err error
			if s, ok := tt.body.(string); ok {
				reqBody = []byte(s)
			} else if tt.body != nil {
				reqBody, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/rebalance", bytes.NewReader(reqBody))
			w := httptest.NewRecorder()

			HandleRebalance(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
