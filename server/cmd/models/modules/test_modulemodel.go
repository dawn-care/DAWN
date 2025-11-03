package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"server/internal/models/modules"
)

func main() {
	err := godotenv.Load("../../../../.env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("✅ Connected to database")

	// Create a module referencing a video metadata ID
	mod := &module.Module{
		ModuleType:    "video",
		CaregiverID:   301,
		CarepartnerID: 201,
		MetadataID:    1,
	}
	if err := module.CreateModule(db, mod); err != nil {
		log.Fatal("Error creating module:", err)
	}
	fmt.Println("Inserted module with ID:", mod.ID)

	// Retrieve module by ID
	got, err := module.GetModuleByID(db, mod.ID)
	if err != nil {
		log.Fatal("Error retrieving module:", err)
	}
	fmt.Printf("Retrieved module: %+v\n", got)

	// Get metadata IDs for caregiver, carepartner, and type
	metadataIDs, err := module.GetMetadataByUsersAndType(db, mod.CaregiverID, mod.CarepartnerID, mod.ModuleType)
	if err != nil {
		log.Fatal("Error getting metadata by user and type:", err)
	}
	fmt.Printf("Metadata IDs for caregiver %d, carepartner %d, type '%s': %v\n",
		mod.CaregiverID, mod.CarepartnerID, mod.ModuleType, metadataIDs)
}
