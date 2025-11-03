package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"server/internal/models/modules"
	"server/internal/models/modules/media"
)

func main() {
	err := godotenv.Load("../../../../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("✅ Connected to database")

	// Create a single photo
	photo := &media.PhotoMedia{
		URI:       "https://example.com/photo1.jpg",
		Filename:  "photo1.jpg",
		Extension: 1,
	}
	err = media.CreatePhoto(db, photo)
	if err != nil {
		log.Fatalf("❌ Error creating photo: %v", err)
	}
	fmt.Printf("✅ Created photo with ID: %d\n", photo.ID)

	// Create a module referencing this photo
	mod := &module.Module{
		ModuleType:    "photo",
		CaregiverID:   1,
		CarepartnerID: 1,
		MetadataID:    photo.ID,
	}
	err = module.CreateModule(db, mod)
	if err != nil {
		log.Fatalf("❌ Error creating module: %v", err)
	}
	fmt.Printf("✅ Created module with ID: %d\n", mod.ID)

	// Retrieve the photo by ID
	retrieved, err := media.GetPhotoByID(db, photo.ID)
	if err != nil {
		log.Fatalf("❌ Error retrieving photo: %v", err)
	}
	fmt.Printf("✅ Retrieved photo: %+v\n", retrieved)

	// Bulk insert photos
	photos := []media.PhotoMedia{
		{URI: "https://example.com/photo2.jpg", Filename: "photo2.jpg", Extension: 1},
		{URI: "https://example.com/photo3.jpg", Filename: "photo3.jpg", Extension: 1},
	}
	err = media.CreatePhotosBulk(db, photos)
	if err != nil {
		log.Fatalf("❌ Error bulk inserting photos: %v", err)
	}
	fmt.Printf("✅ Bulk inserted %d photos\n", len(photos))

	// Update photo
	photo.URI = "https://example.com/photo1_updated.jpg"
	err = media.UpdatePhoto(db, photo)
	if err != nil {
		log.Fatalf("❌ Error updating photo: %v", err)
	}
	fmt.Printf("✅ Updated photo %d\n", photo.ID)

	// Retrieve photos by user (using module link)
	userPhotos, err := media.GetPhotosByUsers(db, 1, 1)
	if err != nil {
		log.Fatalf("❌ Error retrieving photos by user: %v", err)
	}
	fmt.Printf("✅ Retrieved %d photos for carepartner 1 & caregiver 1\n", len(userPhotos))

	// Retrieve photos by day (filename pattern)
	dayPhotos, err := media.GetPhotosByDayAndUsers(db, "photo1", 1, 1)
	if err != nil {
		log.Fatalf("❌ Error retrieving photos by day: %v", err)
	}
	fmt.Printf("✅ Retrieved %d photos by day pattern\n", len(dayPhotos))

	// Get total photo stats
	count, err := media.GetPhotoStats(db)
	if err != nil {
		log.Fatalf("❌ Error getting photo stats: %v", err)
	}
	fmt.Printf("📊 Total photos in DB: %d\n", count)

	// Delete photo
	err = media.DeletePhoto(db, photo.ID)
	if err != nil {
		log.Fatalf("❌ Error deleting photo: %v", err)
	}
	fmt.Printf("✅ Deleted photo with ID %d\n", photo.ID)
}
