package main

import (
	"fmt"
	"log"
	"server/internal/models/modules/prompts"
	"testing"

	_ "github.com/lib/pq"
)

// Uses the same ConnectToDB as caregiver_location_model_test

// Insert single carepartner location
func TestCreateCarepartnerLocation(t *testing.T) {
	// Initialize connection
	db := ConnectToDB(t)
	defer db.Close()

	// Create carepartner location
	location := &prompts.CarepartnerLocationPrompt{
		Location: "Care facility",
		Message:  "You are here.",
		URI:      "C:\\image1.png",
	}
	if err := prompts.CreateCarepartnerLocation(db, location); err != nil {
		log.Fatal("CreateCarepartnerLocation failed:", err)
	}
	fmt.Println("Inserted carepartner location:", location.ID)

	// Clean up
	if err := prompts.DeleteCarepartnerLocation(db, location.ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Read carepartner location by ID
func TestReadCarepartnerLocationByID(t *testing.T) {
	// Initialize connection
	db := ConnectToDB(t)
	defer db.Close()

	// Create carepartner location
	location := &prompts.CarepartnerLocationPrompt{
		Location: "Care facility",
		Message:  "You are here.",
		URI:      "C:\\image1.png",
	}
	if err := prompts.CreateCarepartnerLocation(db, location); err != nil {
		log.Fatal("CreateCarepartnerLocation failed:", err)
	}

	// Read the created carepartner location
	got, err := prompts.GetCarepartnerLocationByID(db, location.ID)
	if err != nil {
		log.Fatal("GetCarepartnerLocationByID failed:", err)
	}
	fmt.Println("Got carepartner location:", got)

	// Clean up
	if err := prompts.DeleteCarepartnerLocation(db, location.ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Update carepartner location
func TestUpdateCarepartnerLocation(t *testing.T) {
	// Initialize connection
	db := ConnectToDB(t)
	defer db.Close()

	// Create carepartner location
	location := &prompts.CarepartnerLocationPrompt{
		Location: "Care facility",
		Message:  "You are here.",
		URI:      "C:\\image1.png",
	}
	if err := prompts.CreateCarepartnerLocation(db, location); err != nil {
		log.Fatal("CreateCarepartnerLocation failed:", err)
	}

	// Update carepartner location
	location.Message = "You chose to be here, your family will visit!"
	if err := prompts.UpdateCarepartnerLocation(db, location); err != nil {
		log.Fatal("UpdateCarepartnerLocation failed:", err)
	}
	fmt.Println("Updated carepartner location:", location.ID)

	// Clean up
	if err := prompts.DeleteCarepartnerLocation(db, location.ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Delete carepartner location
func TestDeleteCarepartnerLocation(t *testing.T) {
	// Initialize connection
	db := ConnectToDB(t)
	defer db.Close()

	// Create carepartner location
	location := &prompts.CarepartnerLocationPrompt{
		Location: "Care facility",
		Message:  "You are here.",
		URI:      "C:\\image1.png",
	}
	if err := prompts.CreateCarepartnerLocation(db, location); err != nil {
		log.Fatal("CreateCarepartnerLocation failed:", err)
	}

	// Delete carepartner location
	if err := prompts.DeleteCarepartnerLocation(db, location.ID); err != nil {
		log.Fatal("DeleteCarepartnerLocation failed:", err)
	}
	fmt.Println("Deleted carepartner location:", location.ID)
}
