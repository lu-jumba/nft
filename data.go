package main

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ContractType struct {
	UUID            string  `gorm:"primaryKey" json:"uuid"`
	ShopType        string  `json:"shop_type"`
	FormulaPerDay   string  `json:"formula_per_day"`
	MaxSumInsured   float32 `json:"max_sum_insured"`
	TheftInsured    bool    `json:"theft_insured"`
	Description     string  `json:"description"`
	Conditions      string  `json:"conditions"`
	Active          bool    `json:"active"`
	MinDurationDays int32   `json:"min_duration_days"`
	MaxDurationDays int32   `json:"max_duration_days"`
}

type User struct {
	Username      string   `gorm:"primaryKey" json:"username"`
	Password      string   `json:"password"`
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	ContractIndex []string `gorm:"-" json:"contracts"` // Handled in application logic
}

type Contract struct {
	UUID             string    `gorm:"primaryKey" json:"uuid"`
	Username         string    `json:"username"`
	Item             Item      `gorm:"embedded" json:"item"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	Void             bool      `json:"void"`
	ContractTypeUUID string    `json:"contract_type_uuid"`
	ClaimIndex       []string     `gorm:"foreignKey:ContractUUID" json:"claim_index,omitempty"`

	//ClaimIndex       []string  `gorm:"-" json:"claim_index,omitempty"` // Handled in application logic
}

type Item struct {
	ID          int32   `json:"id"`
	Brand       string  `json:"brand"`
	Model       string  `json:"model"`
	Price       float32 `json:"price"`
	Description string  `json:"description"`
	SerialNo    string  `json:"serial_no"`
}

type Claim struct {
	UUID          string      `gorm:"primaryKey" json:"uuid"`
	ContractUUID  string      `json:"contract_uuid"`
	Date          time.Time   `json:"date"`
	Description   string      `json:"description"`
	IsTheft       bool        `json:"is_theft"`
	Status        ClaimStatus `json:"status"`
	Reimbursable  float32     `json:"reimbursable"`
	Repaired      bool        `json:"repaired"`
	FileReference string      `json:"file_reference"`
}

type RepairOrder struct {
	ClaimUUID    string `json:"claim_uuid"`
	ContractUUID string `json:"contract_uuid"`
	Item         Item   `json:"item"`
	Ready        bool   `json:"ready"`
}

// ClaimStatus Enum
type ClaimStatus int8

const (
	ClaimStatusUnknown ClaimStatus = iota
	ClaimStatusNew
	ClaimStatusRejected
	ClaimStatusRepair
	ClaimStatusReimbursement
	ClaimStatusTheftConfirmed
)

func (s *ClaimStatus) UnmarshalJSON(b []byte) error {
	var value string
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}
	switch strings.ToUpper(value) {
	case "N":
		*s = ClaimStatusNew
	case "J":
		*s = ClaimStatusRejected
	case "R":
		*s = ClaimStatusRepair
	case "F":
		*s = ClaimStatusReimbursement
	case "P":
		*s = ClaimStatusTheftConfirmed
	default:
		*s = ClaimStatusUnknown
	}
	return nil
}

func (s ClaimStatus) MarshalJSON() ([]byte, error) {
	var value string
	switch s {
	case ClaimStatusNew:
		value = "N"
	case ClaimStatusRejected:
		value = "J"
	case ClaimStatusRepair:
		value = "R"
	case ClaimStatusReimbursement:
		value = "F"
	case ClaimStatusTheftConfirmed:
		value = "P"
	default:
		value = ""
	}
	return json.Marshal(value)
}




func (u *User) Contracts(db *gorm.DB) ([]Contract, error) {
	var contracts []Contract
	err := db.Where("uuid IN ?", u.ContractIndex).Find(&contracts).Error
	if err != nil {
		return nil, err
	}
	return contracts, nil
}

func (c *Contract) Claims(db *gorm.DB) ([]Claim, error) {
	var claims []Claim
	err := db.Where("uuid IN ?", c.ClaimIndex).Find(&claims).Error
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (c *Contract) User(db *gorm.DB) (*User, error) {
	if c.Username == "" {
		return nil, errors.New("invalid username in contract")
	}

	var user User
	err := db.Where("username = ?", c.Username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Claim) Contract(db *gorm.DB) (*Contract, error) {
	if c.ContractUUID == "" {
		return nil, errors.New("contract UUID is missing in claim")
	}

	var contract Contract
	err := db.Where("uuid = ?", c.ContractUUID).First(&contract).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

