package main

import (
	"encoding/json"
	"fmt"
	"io"
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

/*func handleListContractTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse optional JSON input
	var input string
	if r.Body != nil {
		defer r.Body.Close()

		// Read the body if provided
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
			return
		}

		// Assign the body as the `args` input if it’s not empty
		if len(bodyBytes) > 0 {
			input = string(bodyBytes)
		}
	}

	// Call the listContractTypes function
	contractTypes, err := listContractTypes(db, input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list contract types: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the list of contract types
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contractTypes)
}*/

func genericHandler[T any](db *gorm.DB, fn interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate HTTP method
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Parse the body for input arguments (if applicable)
		var input string
		if r.Method == http.MethodPost && r.Body != nil {
			defer r.Body.Close()
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
				return
			}
			input = string(bodyBytes)
		}

		switch typedFn := fn.(type) {
		case func(*gorm.DB, string) (T, error):
			// Call the function expecting (T, error)
			result, err := typedFn(db, input)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error processing request: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		case func(*gorm.DB, string) (bool, error):
			// Call the function expecting (bool, error)
			result, err := typedFn(db, input)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error processing request: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"authenticated": result})

		case func(*gorm.DB) ([]map[string]interface{}, error):
			// Call the function expecting ([]map[string]interface{}, error)
			result, err := typedFn(db)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error processing request: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		case func(*gorm.DB, string) error:
			// Call the function expecting only error
			err := typedFn(db, input)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error processing request: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "Success"})

		default:
			http.Error(w, "Invalid function signature", http.StatusInternalServerError)
		}
	}
}


func init() {
	// Initialize the database connection
	db = connectDatabase()
    //migrateDatabase()

	fmt.Println("Database connected successfully")
}

func main() {
	// Create HTTP routes for all functions
	http.HandleFunc("/contract_types/list", genericHandler[[]ContractType](db, listContractTypes))
    http.HandleFunc("/contract_types/create", genericHandler[struct{}](db, createContractType))
    http.HandleFunc("/contract_types/set_active", genericHandler[struct{}](db, setActiveContractType))
    http.HandleFunc("/contracts/list", genericHandler[[]Contract](db, listContracts))
    http.HandleFunc("/claims/list", genericHandler[[]Claim](db, listClaims))
    http.HandleFunc("/contract/create", genericHandler[*Contract](db, createContract))
	http.HandleFunc("/claims/file", genericHandler[struct{}](db, fileClaim))
	http.HandleFunc("/claims/process", genericHandler[struct{}](db, processClaim))
	http.HandleFunc("/users/authenticate", genericHandler[bool](db, authUser))
	http.HandleFunc("/users/get_info", genericHandler[map[string]string](db, getUser))
	http.HandleFunc("/repair_orders/list", genericHandler[[]map[string]interface{}](db, listRepairOrders))
	http.HandleFunc("/repair_orders/complete", genericHandler[struct{}](db, completeRepairOrder))
	http.HandleFunc("/theft_claims/list", genericHandler[[]map[string]interface{}](db, listTheftClaims))
	http.HandleFunc("/theft_claims/process", genericHandler[struct{}](db, processTheftClaim))

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

