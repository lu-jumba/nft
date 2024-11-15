package main

import (
	"encoding/json"
	"errors"

	//"myproject/models" // Import the data package
	"gorm.io/gorm"
)

func listRepairOrders(db *gorm.DB) ([]map[string]interface{}, error) {
	var repairOrders []RepairOrder

	// Query all repair orders where Ready is false
	err := db.Where("ready = ?", false).Find(&repairOrders).Error
	if err != nil {
		return nil, errors.New("failed to fetch repair orders: " + err.Error())
	}

	// Prepare results
	results := []map[string]interface{}{}
	for _, ro := range repairOrders {

		derivedUUID := ro.ClaimUUID

		result := map[string]interface{}{
			"uuid":          derivedUUID,
			"claim_uuid":    ro.ClaimUUID,
			"contract_uuid": ro.ContractUUID,
			"item":          ro.Item,
		}
		results = append(results, result)
	}

	return results, nil
}

func completeRepairOrder(db *gorm.DB, args string) error {
	// Parse input JSON
	input := struct {
		UUID string `json:"uuid"`
	}{}
	err := json.Unmarshal([]byte(args), &input)
	if err != nil {
		return errors.New("invalid input: " + err.Error())
	}

	// Fetch the repair order
	var repairOrder RepairOrder
	err = db.Where("uuid = ?", input.UUID).First(&repairOrder).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("repair order not found")
	} else if err != nil {
		return errors.New("failed to fetch repair order: " + err.Error())
	}

	// Mark the repair order as ready
	repairOrder.Ready = true
	err = db.Save(&repairOrder).Error
	if err != nil {
		return errors.New("failed to update repair order: " + err.Error())
	}

	// Update the associated claim
	var claim Claim
	err = db.Where("contract_uuid = ? AND uuid = ?", repairOrder.ContractUUID, repairOrder.ClaimUUID).First(&claim).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// No claim found; skip updating claim
		return nil
	} else if err != nil {
		return errors.New("failed to fetch associated claim: " + err.Error())
	}

	claim.Repaired = true
	err = db.Save(&claim).Error
	if err != nil {
		return errors.New("failed to update associated claim: " + err.Error())
	}

	return nil
}
