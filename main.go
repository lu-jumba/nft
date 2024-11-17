package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	//"nft/data"
)

/*
var db *gorm.DB

func connectDatabase() {
    dsn := "host=localhost user=postgres password=yourpassword dbname=contract_management port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
}

func migrateDatabase() {
    err := db.AutoMigrate(&ContractType{})
    if err != nil {
        log.Fatalf("Failed to migrate database: %v", err)
    }
}

func initializeContractTypes(w http.ResponseWriter, r *http.Request) {
    var contractTypes []struct {
        UUID string        `json:"uuid"`
        Type ContractType  `json:"contract_type"`
    }

    err := json.NewDecoder(r.Body).Decode(&contractTypes)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    for _, ct := range contractTypes {
        contractType := ContractType{
            UUID:            ct.UUID,
            ShopType:        ct.Type.ShopType,
            FormulaPerDay:   ct.Type.FormulaPerDay,
            MaxSumInsured:   ct.Type.MaxSumInsured,
            TheftInsured:    ct.Type.TheftInsured,
            Description:     ct.Type.Description,
            Conditions:      ct.Type.Conditions,
            Active:          ct.Type.Active,
            MinDurationDays: ct.Type.MinDurationDays,
            MaxDurationDays: ct.Type.MaxDurationDays,
        }
        result := db.Create(&contractType)
        if result.Error != nil {
            log.Printf("Failed to create contract type with UUID %s: %v", ct.UUID, result.Error)
        }
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Contract types initialized successfully"))
}

func main() {
    connectDatabase()
    migrateDatabase()

    http.HandleFunc("/api/contract-types", initializeContractTypes)
    log.Println("Server running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
*/

// Global Database Connection
var db *gorm.DB

func handleListContractTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Call the listContractTypes function
	contractTypes, err := listContractTypes(db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list contract types: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the list of contract types
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contractTypes)
}

func init() {
	// Initialize the database connection
	db = connectDatabase()
    //migrateDatabase()

	fmt.Println("Database connected successfully")
}

func main() {
	// Create HTTP routes for all functions
	http.HandleFunc("/contract_types/list", handleListContractTypes)
	http.HandleFunc("/contract_types/create", handleCreateContractType)
	http.HandleFunc("/contract_types/set_active", handleSetActiveContractType)
	http.HandleFunc("/contracts/list", handleListContracts)
	http.HandleFunc("/claims/list", handleListClaims)
	http.HandleFunc("/claims/file", handleFileClaim)
	http.HandleFunc("/claims/process", handleProcessClaim)
	http.HandleFunc("/users/authenticate", handleAuthUser)
	http.HandleFunc("/users/get_info", handleGetUser)
	http.HandleFunc("/repair_orders/list", handleListRepairOrders)
	http.HandleFunc("/repair_orders/complete", handleCompleteRepairOrder)
	http.HandleFunc("/theft_claims/list", handleListTheftClaims)
	http.HandleFunc("/theft_claims/process", handleProcessTheftClaim)

	// Start the server
	fmt.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func connectDatabase() *gorm.DB {

    dsn := "host=localhost user=postgres password=yourpassword dbname=contract_management port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
	// Replace with actual database connection logic
	// Example:
	// db, err := gorm.Open(postgres.Open("dsn"), &gorm.Config{})
	// if err != nil {
	//     log.Fatalf("Failed to connect to database: %v", err)
	// }
	// return db
	fmt.Println("Mock database connection initialized")
	return &gorm.DB{}
}

