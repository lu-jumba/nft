package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	//"myproject/models" // Adjust to match your project structure
)

func listTheftClaims(db *gorm.DB) ([]map[string]interface{}, error) {
	// Query all claims marked as theft and with status "New"
	var claims []Claim
	if err := db.Where("is_theft = ? AND status = ?", true, ClaimStatusNew).Find(&claims).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch theft claims: %v", err)
	}

	// Prepare results
	results := []map[string]interface{}{}
	for _, claim := range claims {
		// Fetch the associated contract
		var contract Contract
		if err := db.Where("uuid = ?", claim.ContractUUID).First(&contract).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch contract for claim %s: %v", claim.UUID, err)
		}

		// Fetch the associated user
		var user User
		if err := db.Where("username = ?", contract.Username).First(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch user for contract %s: %v", contract.UUID, err)
		}

		// Construct the result
		result := map[string]interface{}{
			"uuid":          claim.UUID,
			"contract_uuid": claim.ContractUUID,
			"item":          contract.Item,
			"description":   claim.Description,
			"name":          fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		}

		
		results = append(results, result)
	}

	return results, nil
}



func processTheftClaim(db *gorm.DB, args string) error {
	// Parse input arguments
	var dto struct {
		UUID          string `json:"uuid"`
		ContractUUID  string `json:"contract_uuid"`
		IsTheft       bool   `json:"is_theft"`
		FileReference string `json:"file_reference"`
	}
	if err := json.Unmarshal([]byte(args), &dto); err != nil {
		return fmt.Errorf("invalid input: %v", err)
	}

	// Fetch the claim
	var claim Claim
	if err := db.Where("uuid = ? AND contract_uuid = ?", dto.UUID, dto.ContractUUID).First(&claim).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("claim not found for UUID: %s", dto.UUID)
		}
		return fmt.Errorf("failed to fetch claim: %v", err)
	}

	// Check if the claim is theft-related and in a valid state
	if !claim.IsTheft || claim.Status != ClaimStatusNew {
		return errors.New("claim is either not related to theft or has an invalid status")
	}

	// Update the claim's status based on theft confirmation
	if dto.IsTheft {
		claim.Status = ClaimStatusTheftConfirmed
	} else {
		claim.Status = ClaimStatusRejected
	}
	claim.FileReference = dto.FileReference

	// Save the updated claim
	if err := db.Save(&claim).Error; err != nil {
		return fmt.Errorf("failed to update claim: %v", err)
	}

	return nil
}

