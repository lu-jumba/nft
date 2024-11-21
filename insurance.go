package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	//"golang.org/x/crypto/bcrypt"

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

	
	// Create the claim
	claim := Claim{
		UUID:         dto.UUID,
		ContractUUID: dto.ContractUUID,
		Date:         dto.Date,
		Description:  dto.Description,
		IsTheft:      dto.IsTheft,
		Status:       ClaimStatusNew,
	}

	// Check if the contract exists
	var contract Contract
	if err := db.Where("uuid = ?", dto.ContractUUID).First(&claim).Error; err != nil {  //was First(&contract) 
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("contract not found: %s", dto.ContractUUID)
		}
		return fmt.Errorf("failed to fetch contract: %v", err)
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


func processClaim(db *gorm.DB, args string) error {
	// Parse input arguments
	var input struct {
		UUID         string              `json:"uuid"`
		ContractUUID string              `json:"contract_uuid"`
		Status       ClaimStatus  `json:"status"`
		Reimbursable float32             `json:"reimbursable"`
	}
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return fmt.Errorf("invalid input: %v", err)
	}

	// Fetch the claim
	var claim Claim
	if err := db.Where("uuid = ? AND contract_uuid = ?", input.UUID, input.ContractUUID).First(&claim).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("claim not found for UUID: %s", input.UUID)
		}
		return fmt.Errorf("failed to fetch claim: %v", err)
	}

	// Validation logic
	if !claim.IsTheft && claim.Status != ClaimStatusNew {
		return errors.New("cannot change the status of a non-new claim")
	}
	if claim.IsTheft && claim.Status == ClaimStatusNew {
		return errors.New("theft must first be confirmed by authorities")
	}

	// Update the claim status
	claim.Status = input.Status
	switch input.Status {
	case ClaimStatusRepair:
		// Approve repair
		if claim.IsTheft {
			return errors.New("cannot repair stolen items")
		}
		claim.Reimbursable = 0

		// Create a repair order
		var contract Contract
		if err := db.Where("uuid = ?", input.ContractUUID).First(&claim).Error; err != nil {  //was First(&contract)
			return fmt.Errorf("contract not found for UUID: %s", input.ContractUUID)
		}

		repairOrder := RepairOrder{
			Item:         contract.Item,
			ClaimUUID:    input.UUID,
			ContractUUID: input.ContractUUID,
			Ready:        false,
		}
		if err := db.Create(&repairOrder).Error; err != nil {
			return fmt.Errorf("failed to create repair order: %v", err)
		}

	case ClaimStatusReimbursement:
		// Approve reimbursement
		claim.Reimbursable = input.Reimbursable
		if claim.IsTheft {
			// Fetch and void the associated contract
			var contract Contract
			if err := db.Where("uuid = ?", input.ContractUUID).First(&claim).Error; err != nil { //was First(&contract)
				return fmt.Errorf("contract not found for UUID: %s", input.ContractUUID)
			}
			contract.Void = true
			if err := db.Save(&contract).Error; err != nil {
				return fmt.Errorf("failed to update contract: %v", err)
			}
		}

	case ClaimStatusRejected:
		// Mark the claim as rejected
		claim.Reimbursable = 0

	default:
		return errors.New("unknown status change")
	}

	// Save the updated claim
	if err := db.Save(&claim).Error; err != nil {
		return fmt.Errorf("failed to update claim: %v", err)
	}

	return nil
}



func authUser(db *gorm.DB, args string) (bool, error) {
	// Parse input arguments
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return false, fmt.Errorf("invalid input: %v", err)
	}

	// Fetch the user from the database
	var user User
	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // User not found, authentication fails
		}
		return false, fmt.Errorf("failed to fetch user: %v", err)
	}

	// Verify the password
	authenticated := user.Password == input.Password

	return authenticated, nil
}


func getUser(db *gorm.DB, args string) (map[string]string, error) {
	// Parse input arguments
	var input struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return nil, fmt.Errorf("invalid input: %v", err)
	}

	// Fetch the user from the database
	var user User
	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found, return nil map
		}
		return nil, fmt.Errorf("failed to fetch user: %v", err)
	}

	// Construct the response
	response := map[string]string{
		"username":  user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	}

	return response, nil
}


// UpdatePassword updates a user's password based on their username.
func updatePassword(db *gorm.DB, args string) (bool, error) {
	// Parse input arguments
	var input struct {
		Username    string `json:"username"`
		NewPassword string `json:"new_password"`
	}
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return false, fmt.Errorf("invalid input: %v", err)
	}

	// Validate input
	if input.Username == "" || input.NewPassword == "" {
		return false, errors.New("username and new password must not be empty")
	}

	// Fetch the user from the database
	var user User
	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("user not found")
		}
		return false, fmt.Errorf("failed to fetch user: %v", err)
	}

	// Hash the new password
	hashedPassword, err := HashPassword(input.NewPassword)
	if err != nil {
		return false, fmt.Errorf("failed to hash password: %v", err)
	}

	// Update the password in the database
	if err := db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return false, fmt.Errorf("failed to update password: %v", err)
	}

	return true, nil
}

/*
type User struct {
    Username        string    `gorm:"primaryKey" json:"username"`
    Password        string    `json:"password"`
    Email           string    `json:"email"`
    FailedAttempts  int       `json:"failed_attempts"`
    LastFailedLogin time.Time `json:"last_failed_login"`
}


package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"myproject/models" // Adjust to match your project structure
)

const lockoutDuration = 2 * time.Minute
const maxFailedAttempts = 3

func authUser(db *gorm.DB, args string) (bool, error) {
	// Parse input arguments
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return false, fmt.Errorf("invalid input: %v", err)
	}

	// Fetch the user from the database
	var user models.User
	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // User not found, authentication fails
		}
		return false, fmt.Errorf("failed to fetch user: %v", err)
	}

	// Check if the user is currently locked out
	if user.FailedAttempts >= maxFailedAttempts {
		timeSinceLastFailed := time.Since(user.LastFailedLogin)
		if timeSinceLastFailed < lockoutDuration {
			return false, fmt.Errorf("account locked. Try again in %v", lockoutDuration-timeSinceLastFailed)
		}

		// Lockout period has expired; reset failed attempts
		user.FailedAttempts = 0
		if err := db.Save(&user).Error; err != nil {
			return false, fmt.Errorf("failed to reset failed attempts: %v", err)
		}
	}

	// Verify the password
	authenticated := user.Password == input.Password // Replace with hashed password comparison in production
	if !authenticated {
		// Increment failed attempts
		user.FailedAttempts++
		user.LastFailedLogin = time.Now()
		if err := db.Save(&user).Error; err != nil {
			return false, fmt.Errorf("failed to update failed attempts: %v", err)
		}

		if user.FailedAttempts >= maxFailedAttempts {
			return false, errors.New("account locked due to multiple failed attempts")
		}

		return false, errors.New("invalid username or password")
	}

	// Reset failed attempts on successful login
	user.FailedAttempts = 0
	if err := db.Save(&user).Error; err != nil {
		return false, fmt.Errorf("failed to reset failed attempts: %v", err)
	}

	return true, nil
}

//Password Reset Functionality
//If the user is locked out, send a password reset link:

//Generate a Password Reset Token:

//Create a secure token and save it in a PasswordResetTokens table along with its expiration time.
//Send Reset Email:

//Use an email service to send the reset link to the user's registered email address.
func sendPasswordResetEmail(db *gorm.DB, email string) error {
	// Fetch the user by email
	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no user found with the given email")
		}
		return fmt.Errorf("failed to fetch user: %v", err)
	}

	// Generate a secure reset token
	resetToken := generateSecureToken()
	resetExpiry := time.Now().Add(15 * time.Minute)

	// Save the token in a password reset table
	passwordReset := models.PasswordReset{
		UserID:  user.Username,
		Token:   resetToken,
		Expires: resetExpiry,
	}
	if err := db.Create(&passwordReset).Error; err != nil {
		return fmt.Errorf("failed to save password reset token: %v", err)
	}

	// Send the email (use an email service)
	emailBody := fmt.Sprintf("Click the link to reset your password: https://example.com/reset?token=%s", resetToken)
	if err := sendEmail(user.Email, "Password Reset Request", emailBody); err != nil {
		return fmt.Errorf("failed to send password reset email: %v", err)
	}

	return nil
}*/

