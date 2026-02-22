package services

import (
	"portfolio-rebalancer/internal/models"
)

func CalculateRebalance(userID string, newAllocation, currentAllocation map[string]float64) []models.RebalanceTransaction {
	var result []models.RebalanceTransaction

	// Identify all unique assets
	assets := make(map[string]bool)
	for k := range newAllocation {
		assets[k] = true
	}
	for k := range currentAllocation {
		assets[k] = true
	}

	for asset := range assets {
		marketPct := newAllocation[asset]
		targetPct := currentAllocation[asset]
		diff := targetPct - marketPct

		if diff == 0 {
			continue
		}

		tx := models.RebalanceTransaction{
			UserID: userID,
			Asset:  asset,
		}

		if diff > 0 {
			tx.Action = "BUY"
			tx.RebalancePercent = diff
		} else {
			tx.Action = "SELL"
			tx.RebalancePercent = -diff
		}

		result = append(result, tx)
	}

	return result
}
