package main

import (
	"fmt"
	"os"

	"github.com/go-chi/chi/v5/middleware"

	"database/sql"
	"log"
	"net/http"
	"time"

	mediaCtrlr "server/internal/controllers/media"
	caregiverCtrlr "server/internal/controllers/user/caregiver"
	mediaRts "server/internal/routes/media"
	caregiverRts "server/internal/routes/user"
	mediaSvc "server/internal/services/media"
	caregiverSvc "server/internal/services/user/caregiver"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq" // or whatever driver you use
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	// Router Setup
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Example
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World!")
	})

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	// Actual
	// DB SETUP
	db, dberr := sql.Open("postgres", "postgres://user:password@localhost:5432/dbname?sslmode=disable")
	if dberr != nil {
		log.Fatal("failed to connect to DB:", dberr)
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	// MINIO SETUP
	minioClient, bloberr := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minio-access-key", "minio-secret-key", ""),
		Secure: false, // true if using https
	})
	if bloberr != nil {
		log.Fatal("failed to connect to MinIO:", bloberr)
	}

	// INITIALIZE SERVICES
	// Media
	audioService := mediaSvc.NewAudioService(db, minioClient, "audio-bucket")
	photoService := mediaSvc.NewPhotoService(db, minioClient, "photo-bucket")
	videoService := mediaSvc.NewVideoService(db, minioClient, "video-bucket")
	// User
	caregiverService := caregiverSvc.NewCaregiverService(db)

	// INITIALIZE CTRLRS
	// Media
	audioCtrl := mediaCtrlr.NewAudioController(audioService)
	photoCtrl := mediaCtrlr.NewPhotoController(photoService)
	videoCtrl := mediaCtrlr.NewVideoController(videoService)
	// User
	caregiverCtrl := caregiverCtrlr.NewCaregiverController(caregiverService)

	// ROUTE INITIALIZATION
	// Media
	mediaRts.RegisterAudioRoutes(r, audioCtrl)
	mediaRts.RegisterPhotoRoutes(r, photoCtrl)
	mediaRts.RegisterVideoRoutes(r, videoCtrl)
	// User
	caregiverRts.RegisterCaregiverRoutes(r, caregiverCtrl)

	// Launch
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting on port %s\n", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
