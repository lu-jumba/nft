package main

import (
    "encoding/json"
    "log"
    "net/http"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
	//"nft/data"
)




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
