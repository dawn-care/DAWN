package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"server/internal/models/modules"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("../../../../.env")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a module referencing a video metadata ID
	mod := &module.Module{ModuleType: "video", CaregiverID: 301, CarepartnerID: 201, MetadataID: 1}
	if err := module.CreateModule(db, mod); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted module:", mod.ID)

	// Retrieve module by ID
	got, _ := module.GetModuleByID(db, mod.ID)
	fmt.Println("Got module:", got)

	// Get the owners of a video through metadata ID
	cpID, cgID, _ := module.GetVideoOwners(db, mod.MetadataID)
	fmt.Printf("Owners of video %d -> Carepartner: %d, Caregiver: %d\n", mod.MetadataID, cpID, cgID)
}
