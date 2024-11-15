package main

import (
	"encoding/json"
	"errors"
	"time"

	//"myproject/models" // Import the data package
	"gorm.io/gorm"
)

func createContract(db *gorm.DB, args string) (*Contract, error) {
	// Parse the input JSON
	dto := struct {
		UUID             string        `json:"uuid"`
		ContractTypeUUID string        `json:"contract_type_uuid"`
		Username         string        `json:"username"`
		Password         string        `json:"password"`
		FirstName        string        `json:"first_name"`
		LastName         string        `json:"last_name"`
		Item             Item          `json:"item"`
		StartDate        time.Time     `json:"start_date"`
		EndDate          time.Time     `json:"end_date"`
	}{}

	err := json.Unmarshal([]byte(args), &dto)
	if err != nil {
		return nil, errors.New("invalid input: " + err.Error())
	}

	// Check if user exists or create a new one
	var user User
	err = db.Where("username = ?", dto.Username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) && dto.Password != "" {
		// Create a new user
		user = User{
			Username:  dto.Username,
			Password:  dto.Password,
			FirstName: dto.FirstName,
			LastName:  dto.LastName,
		}
		if err := db.Create(&user).Error; err != nil {
			return nil, errors.New("failed to create user: " + err.Error())
		}
	} else if err != nil {
		return nil, errors.New("failed to query user: " + err.Error())
	}

	// Check if the contract type exists
	var contractType ContractType
	err = db.Where("uuid = ?", dto.ContractTypeUUID).First(&contractType).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("contract type not found")
	} else if err != nil {
		return nil, errors.New("failed to query contract type: " + err.Error())
	}

	// Create the contract
	contract := &Contract{
		UUID:             dto.UUID,
		Username:         dto.Username,
		ContractTypeUUID: dto.ContractTypeUUID,
		Item:             dto.Item,
		StartDate:        dto.StartDate,
		EndDate:          dto.EndDate,
		Void:             false,
		ClaimIndex:       []string{},
	}

	if err := db.Create(contract).Error; err != nil {
		return nil, errors.New("failed to create contract: " + err.Error())
	}

	return contract, nil
}


func createUser(db *gorm.DB, args string) (*User, error) {
	// Parse the input JSON
	var user User
	err := json.Unmarshal([]byte(args), &user)
	if err != nil {
		return nil, errors.New("invalid input: " + err.Error())
	}

	// Check if the user already exists
	var existingUser User
	err = db.Where("username = ?", user.Username).First(&existingUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// User does not exist, create a new one
		if err := db.Create(&user).Error; err != nil {
			return nil, errors.New("failed to create user: " + err.Error())
		}
		return &user, nil
	} else if err != nil {
		return nil, errors.New("failed to query user: " + err.Error())
	}

	// User already exists, return the existing user
	return &existingUser, nil
}
