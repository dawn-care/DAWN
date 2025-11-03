package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"server/internal/models/modules/prompts"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func TestCarepartnerLocationModel() {
	godotenv.Load("../../../../.env")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Insert single carepartner location
	location := &prompts.CarepartnerLocationPrompt{
		Location: "1 Lomb Memorial Driver, Rochester NY, 14623",
		Message:  "You are at this facility.",
		URI:      "C:\\image1.png",
	}
	if err := prompts.CreateCarepartnerLocation(db, location); err != nil {
		log.Fatal("CreateCarepartnerLocation failed:", err)
	}
	fmt.Println("Inserted carepartner location:", location.ID)

	// Read carepartner location by ID
	got, err := prompts.GetCarepartnerLocationByID(db, location.ID)
	if err != nil {
		log.Fatal("GetCarepartnerLocationByID failed:", err)
	}
	fmt.Println("Got carepartner location:", got)

	// Update carepartner location
	location.Message = "You are at this facility. Here is a picture of you signing up to be here!"
	if err := prompts.UpdateCarepartnerLocation(db, location); err != nil {
		log.Fatal("UpdateCarepartnerLocation failed:", err)
	}
	fmt.Println("Updated carepartner location:", location.ID)

	// Delete carepartner location
	if err := prompts.DeleteCarepartnerLocation(db, location.ID); err != nil {
		log.Fatal("DeleteCarepartnerLocation failed:", err)
	}
	fmt.Println("Deleted carepartner location:", location.ID)
}
