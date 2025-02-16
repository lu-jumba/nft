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

func genericHandler[T any](db *gorm.DB, fn interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle OPTIONS request for preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var input string

		// Parse request body for non-GET methods
		if r.Method != http.MethodGet && r.Body != nil {
			defer r.Body.Close()
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
				return
			}
			input = string(bodyBytes)
		}

		// Dynamic function handling
		switch typedFn := fn.(type) {
		case func(*gorm.DB, string) (T, error):
			// Function expects (T, error)
			result, err := typedFn(db, input)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		case func(*gorm.DB) ([]T, error):
			// Function expects ([]T, error)
			result, err := typedFn(db)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		case func(*gorm.DB, string) error:
			// Function expects only error
			err := typedFn(db, input)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "Success"})

		default:
			http.Error(w, `{"error": "Invalid function signature"}`, http.StatusInternalServerError)
		}
	}
}


func init() {
	// Initialize the database connection
	db = connectDatabase()
    
	// Perform auto-migration
	migrateDatabase(db)

	fmt.Println("Database connected successfully")
}

func main() {
	// Create HTTP routes for all functions
	http.HandleFunc("/contract_type_ls", genericHandler[[]ContractType](db, listContractTypes))
    http.HandleFunc("/contract_type_create", genericHandler[struct{}](db, createContractType))
    http.HandleFunc("/contract_type_set_active", genericHandler[struct{}](db, setActiveContractType))
    http.HandleFunc("/contract_ls", genericHandler[[]Contract](db, listContracts))
    http.HandleFunc("/claim_ls", genericHandler[[]Claim](db, listClaims))
    http.HandleFunc("/contract_create", genericHandler[*Contract](db, createContract))
	http.HandleFunc("/claim_file", genericHandler[struct{}](db, fileClaim))
	http.HandleFunc("/claim_process", genericHandler[struct{}](db, processClaim))
	http.HandleFunc("/user_authenticate", genericHandler[bool](db, authUser))
	http.HandleFunc("/user_get_info", genericHandler[map[string]string](db, getUser))
	http.HandleFunc("/repair_order_ls", genericHandler[[]map[string]interface{}](db, listRepairOrders))
	http.HandleFunc("/repair_order_complete", genericHandler[struct{}](db, completeRepairOrder))
	http.HandleFunc("/theft_claim_ls", genericHandler[[]map[string]interface{}](db, listTheftClaims))
	http.HandleFunc("/theft_claim_process", genericHandler[struct{}](db, processTheftClaim))

	// Start the server
	fmt.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))

	/*http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			genericHandler[map[string]string](db, getUser)(w, r)
		case http.MethodPost:
			genericHandler[bool](db, authUser)(w, r)
		case http.MethodPut:
			genericHandler[bool](db, updatePassword)(w, r) // Assuming you have this function
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})*/
	
}

func connectDatabase() *gorm.DB {

    dsn := "host=localhost user=postgres password=yourpassword dbname=contract_management port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
	
	// return db
	fmt.Println("Mock database connection initialized")
	return &gorm.DB{}
}


func migrateDatabase(db *gorm.DB) {
	err := db.AutoMigrate(&User{}, &ContractType{}, &Contract{}, &Claim{}, &RepairOrder{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migrated successfully")
}

