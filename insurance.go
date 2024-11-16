package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func listContractTypes(db *gorm.DB, args string) ([]ContractType, error) {
	// Check if the request is from a merchant
	callingAsMerchant := len(args) > 0
	var input struct {
		ShopType string `json:"shop_type"`
	}

	// Parse input arguments for merchant filtering
	if callingAsMerchant {
		if err := json.Unmarshal([]byte(args), &input); err != nil {
			return nil, fmt.Errorf("invalid input: %v", err)
		}
	}

	// Query contract types
	var contractTypes []ContractType
	query := db.Model(&ContractType{})

	// Apply filtering if the request is from a merchant
	if callingAsMerchant {
		query = query.Where("active = ? AND shop_type ILIKE ?", true, "%"+strings.ToTitle(input.ShopType)+"%")
	}

	// Execute the query
	if err := query.Find(&contractTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch contract types: %v", err)
	}

	return contractTypes, nil
}


func createContractType(db *gorm.DB, args string) error {
	// Parse input
	var partial struct {
		UUID string `json:"uuid"`
	}
	var contractType ContractType

	// Extract UUID
	if err := json.Unmarshal([]byte(args), &partial); err != nil {
		return fmt.Errorf("invalid input: %v", err)
	}

	// Parse full contract type
	if err := json.Unmarshal([]byte(args), &contractType); err != nil {
		return fmt.Errorf("invalid input: %v", err)
	}

	// Assign UUID to the contract type
	contractType.UUID = partial.UUID

	// Save to the database
	if err := db.Create(&contractType).Error; err != nil {
		return fmt.Errorf("failed to create contract type: %v", err)
	}

	return nil
}

func setActiveContractType(db *gorm.DB, args string) error {
	// Parse input
	var input struct {
		UUID   string `json:"uuid"`
		Active bool   `json:"active"`
	}

	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return fmt.Errorf("invalid input: %v", err)
	}

	// Fetch the contract type
	var contractType ContractType
	if err := db.Where("uuid = ?", input.UUID).First(&contractType).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("contract type with UUID %s not found", input.UUID)
		}
		return fmt.Errorf("failed to query contract type: %v", err)
	}

	// Update the active status
	contractType.Active = input.Active
	if err := db.Save(&contractType).Error; err != nil {
		return fmt.Errorf("failed to update contract type: %v", err)
	}

	return nil
}


func listContracts(db *gorm.DB, args string) ([]Contract, error) {
	// Parse input arguments for optional username filtering
	var input struct {
		Username string `json:"username"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal([]byte(args), &input); err != nil {
			return nil, fmt.Errorf("invalid input: %v", err)
		}
	}
	filterByUsername := len(input.Username) > 0

	// Query contracts with claims preloaded
	var contracts []Contract
	query := db.Model(&Contract{}).Preload("Claims")
	if filterByUsername {
		query = query.Where("username = ?", input.Username)
	}

	// Execute the query
	if err := query.Find(&contracts).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch contracts: %v", err)
	}

	// Return the contracts with preloaded claims
	return contracts, nil
}


func listClaims(db *gorm.DB, args string) ([]Claim, error) {
	// Parse input arguments for optional status filtering
	var input struct {
		Status ClaimStatus `json:"status"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal([]byte(args), &input); err != nil {
			return nil, fmt.Errorf("invalid input: %v", err)
		}
	}

	// Query claims with optional status filtering
	var claims []Claim
	query := db.Model(&Claim{})
	if input.Status != ClaimStatusUnknown {
		query = query.Where("status = ?", input.Status)
	}

	if err := query.Find(&claims).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch claims: %v", err)
	}

	return claims, nil
}


func fileClaim(db *gorm.DB, args string) error {
	// Parse input arguments
	var dto struct {
		UUID         string    `json:"uuid"`
		ContractUUID string    `json:"contract_uuid"`
		Date         time.Time `json:"date"`
		Description  string    `json:"description"`
		IsTheft      bool      `json:"is_theft"`
	}
	if err := json.Unmarshal([]byte(args), &dto); err != nil {
		return fmt.Errorf("invalid input: %v", err)
	}

	// Check if the contract exists
	var contract Contract
	if err := db.Where("uuid = ?", dto.ContractUUID).First(&contract).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("contract not found: %s", dto.ContractUUID)
		}
		return fmt.Errorf("failed to fetch contract: %v", err)
	}

	// Create the claim
	claim := Claim{
		UUID:         dto.UUID,
		ContractUUID: dto.ContractUUID,
		Date:         dto.Date,
		Description:  dto.Description,
		IsTheft:      dto.IsTheft,
		Status:       ClaimStatusNew,
	}

	// Save the claim to the database
	if err := db.Create(&claim).Error; err != nil {
		return fmt.Errorf("failed to file claim: %v", err)
	}

	// Update the claim index in the contract (if needed)
	contract.ClaimIndex = append(contract.ClaimIndex, claim.UUID)
	if err := db.Save(&contract).Error; err != nil {
		return fmt.Errorf("failed to update contract claim index: %v", err)
	}

	return nil
}


