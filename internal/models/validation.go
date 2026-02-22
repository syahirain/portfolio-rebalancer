package models

import (
	"errors"
)

// ValidateUserAndAllocation checks UserID and Allocation map validity
func ValidateUserAndAllocation(userID string, allocation map[string]float64) error {
	if userID == "" {
		return errors.New("userId is required")
	}
	if len(allocation) == 0 {
		return errors.New("allocation is required")
	}

	var total float64
	for _, pct := range allocation {
		if pct < 0 {
			return errors.New("allocation percentages cannot be negative")
		}
		total += pct
	}

	if total != 100 {
		return errors.New("allocation percentages must sum to 100")
	}

	return nil
}

func ValidatePercentage(allocation map[string]float64) error {
	var totalPercentage float64
	for _, percentage := range allocation {
		if percentage < 0 {
			return errors.New("percentage cannot be negative")
		}
		totalPercentage += percentage
	}

	if totalPercentage != 100 {
		return errors.New("total allocation percentage must be 100%")
	}

	return nil
}
