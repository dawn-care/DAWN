package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"server/internal/models/modules/media"
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

	// Insert single video
	video1 := &media.VideoMedia{URI: "C:\\video1.mp4", Filename: "video1.mp4", Extension: 1}
	if err := media.CreateVideo(db, video1); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted video:", video1.ID)

	// Insert multiple videos
	videos := []media.VideoMedia{
		{URI: "C:\\video2.mp4", Filename: "video2.mp4", Extension: 1},
		{URI: "C:\\video3.mp4", Filename: "video3.mp4", Extension: 2},
	}
	if err := media.CreateVideosBulk(db, videos); err != nil {
		log.Fatal(err)
	}
	for _, v := range videos {
		fmt.Println("Bulk inserted video:", v.ID)
	}

	// Retrieve video by ID
	got, _ := media.GetVideoByID(db, video1.ID)
	fmt.Println("Got video:", got)

	// Update video
	video1.Filename = "video1_updated.mp4"
	if err := media.UpdateVideo(db, video1); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated video:", video1.ID)

	// Delete a video
	if err := media.DeleteVideo(db, videos[1].ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted video:", videos[1].ID)

	// Get video count
	count, _ := media.GetVideoStats(db)
	fmt.Println("Total videos:", count)
}
