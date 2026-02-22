package models

type Portfolio struct {
	UserID     string             `json:"user_id"`
	Allocation map[string]float64 `json:"allocation"` // Current user allocation in percentage terms
}

type UpdatedPortfolio struct {
	UserID        string             `json:"user_id"`
	NewAllocation map[string]float64 `json:"new_allocation"` // Updated user allocation from provider in percentage terms
}

type RebalancePortfolioKafka struct {
	UserID            string             `json:"user_id"`
	NewAllocation     map[string]float64 `json:"new_allocation"`     // Updated user allocation from provider in percentage terms
	CurrentAllocation map[string]float64 `json:"current_allocation"` // Current user allocation in percentage terms
}

type RebalanceTransaction struct {
	UserID           string  `json:"user_id"`
	Action           string  `json:"action"`            // "BUY" or "SELL"
	Asset            string  `json:"asset"`             // "stocks", "bonds", "gold"
	RebalancePercent float64 `json:"rebalance_percent"` // percentage to buy/sell
}

type RebalanceRequest struct {
	UserID         string `json:"user_id"`
	AllocationHash string `json:"allocation_hash"` // Hash of the updated allocation json from the provider
	Status         string `json:"status"`          // "PENDING", "COMPLETED", "FAILED"
}
