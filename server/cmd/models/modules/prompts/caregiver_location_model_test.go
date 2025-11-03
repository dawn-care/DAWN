package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"server/internal/models/modules/prompts"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Helper function to connect to the database
func ConnectToDB(t *testing.T) *sql.DB {
	godotenv.Load("../../../../../.env")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Insert single caregiver location
func TestCreateCaregiverLocation(t *testing.T) {
	// Initialize connection
	db := ConnectToDB(t)
	defer db.Close()

	// Create caregiver location
	location := &prompts.CaregiverLocationPrompt{
		Location: "1 Lomb Memorial Driver, Rochester NY, 14623",
		Message:  "Helena is only 14 minutes away!",
		URI:      "C:\\image1.png",
	}
	if err := prompts.CreateCaregiverLocation(db, location); err != nil {
		log.Fatal("CreateCaregiverLocation failed:", err)
	}
	fmt.Println("Inserted caregiver location:", location.ID)

	// Clean up
	if err := prompts.DeleteCaregiverLocation(db, location.ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Read caregiver location by ID
func TestReadCaregiverLocationByID(t *testing.T) {
	// Initialize connection
	db := ConnectToDB(t)
	defer db.Close()

	// Create caregiver location
	location := &prompts.CaregiverLocationPrompt{
		Location: "1 Lomb Memorial Driver, Rochester NY, 14623",
		Message:  "Helena is only 14 minutes away!",
		URI:      "C:\\image1.png",
	}
	if err := prompts.CreateCaregiverLocation(db, location); err != nil {
		log.Fatal("CreateCaregiverLocation failed:", err)
	}

	// Read the created caregiver location
	got, err := prompts.GetCaregiverLocationByID(db, location.ID)
	if err != nil {
		log.Fatal("GetCaregiverLocationByID failed:", err)
	}
	fmt.Println("Got caregiver location:", got)

	// Clean up
	if err := prompts.DeleteCaregiverLocation(db, location.ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Update caregiver location
func TestUpdateCaregiverLocation(t *testing.T) {
	// Initialize connection
	db := ConnectToDB(t)
	defer db.Close()

	// Create caregiver location
	location := &prompts.CaregiverLocationPrompt{
		Location: "1 Lomb Memorial Driver, Rochester NY, 14623",
		Message:  "Helena is only 14 minutes away!",
		URI:      "C:\\image1.png",
	}
	if err := prompts.CreateCaregiverLocation(db, location); err != nil {
		log.Fatal("CreateCaregiverLocation failed:", err)
	}

	// Update caregiver location
	location.Message = "Helena is only 10 minutes away!"
	if err := prompts.UpdateCaregiverLocation(db, location); err != nil {
		log.Fatal("UpdateCaregiverLocation failed:", err)
	}
	fmt.Println("Updated caregiver location:", location.ID)

	// Clean up
	if err := prompts.DeleteCaregiverLocation(db, location.ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Delete caregiver location
func TestDeleteCaregiverLocation(t *testing.T) {
	// Initialize connection
	db := ConnectToDB(t)
	defer db.Close()

	// Create caregiver location
	location := &prompts.CaregiverLocationPrompt{
		Location: "1 Lomb Memorial Driver, Rochester NY, 14623",
		Message:  "Helena is only 14 minutes away!",
		URI:      "C:\\image1.png",
	}
	if err := prompts.CreateCaregiverLocation(db, location); err != nil {
		log.Fatal("CreateCaregiverLocation failed:", err)
	}

	// Delete caregiver location
	if err := prompts.DeleteCaregiverLocation(db, location.ID); err != nil {
		log.Fatal("DeleteCaregiverLocation failed:", err)
	}
	fmt.Println("Deleted caregiver location:", location.ID)
}
