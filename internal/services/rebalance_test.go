package services

import (
	"portfolio-rebalancer/internal/models"
	"reflect"
	"sort"
	"testing"
)

func TestCalculateRebalance(t *testing.T) {
	tests := []struct {
		name              string
		userID            string
		newAllocation     map[string]float64
		currentAllocation map[string]float64
		expected          []models.RebalanceTransaction
	}{
		{
			name:   "Buy action needed",
			userID: "user1",
			newAllocation: map[string]float64{
				"Stocks": 40.0,
			},
			currentAllocation: map[string]float64{
				"Stocks": 50.0,
			},
			expected: []models.RebalanceTransaction{
				{
					UserID:           "user1",
					Asset:            "Stocks",
					Action:           "BUY",
					RebalancePercent: 10.0,
				},
			},
		},
		{
			name:   "Sell action needed",
			userID: "user1",
			newAllocation: map[string]float64{
				"Bonds": 60.0,
			},
			currentAllocation: map[string]float64{
				"Bonds": 50.0,
			},
			expected: []models.RebalanceTransaction{
				{
					UserID:           "user1",
					Asset:            "Bonds",
					Action:           "SELL",
					RebalancePercent: 10.0,
				},
			},
		},
		{
			name:   "No action needed",
			userID: "user1",
			newAllocation: map[string]float64{
				"Cash": 10.0,
			},
			currentAllocation: map[string]float64{
				"Cash": 10.0,
			},
			expected: nil,
		},
		{
			name:   "Mixed actions",
			userID: "user1",
			newAllocation: map[string]float64{
				"Stocks": 60.0,
				"Bonds":  40.0,
			},
			currentAllocation: map[string]float64{
				"Stocks": 50.0,
				"Bonds":  50.0,
			},
			expected: []models.RebalanceTransaction{
				{
					UserID:           "user1",
					Asset:            "Stocks",
					Action:           "SELL",
					RebalancePercent: 10.0,
				},
				{
					UserID:           "user1",
					Asset:            "Bonds",
					Action:           "BUY",
					RebalancePercent: 10.0,
				},
			},
		},
		{
			name:   "Asset missing in new allocation (implied 0)",
			userID: "user1",
			newAllocation: map[string]float64{},
			currentAllocation: map[string]float64{
				"Gold": 10.0,
			},
			expected: []models.RebalanceTransaction{
				{
					UserID:           "user1",
					Asset:            "Gold",
					Action:           "BUY",
					RebalancePercent: 10.0,
				},
			},
		},
		{
			name:   "Asset missing in current allocation (implied 0)",
			userID: "user1",
			newAllocation: map[string]float64{
				"Gold": 10.0,
			},
			currentAllocation: map[string]float64{},
			expected: []models.RebalanceTransaction{
				{
					UserID:           "user1",
					Asset:            "Gold",
					Action:           "SELL",
					RebalancePercent: 10.0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRebalance(tt.userID, tt.newAllocation, tt.currentAllocation)

			// Sort results to ensure deterministic comparison
			sortTransactions(result)
			sortTransactions(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("CalculateRebalance() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func sortTransactions(txs []models.RebalanceTransaction) {
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Asset < txs[j].Asset
	})
}