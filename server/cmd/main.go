package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	mediaCtrlr "server/internal/controllers/media"
	modulesCtrlr "server/internal/controllers/modules"
	promptsCtrlr "server/internal/controllers/prompts"
	caregiverCtrlr "server/internal/controllers/user/caregiver"
	newsCtrlr "server/internal/controllers/widgets/news"
	weatherCtrlr "server/internal/controllers/widgets/weather"

	mediaRts "server/internal/routes/media"
	modulesRts "server/internal/routes/modules"
	promptsRts "server/internal/routes/prompts"
	caregiverRts "server/internal/routes/user"
	newsRts "server/internal/routes/widgets/news"
	weatherRts "server/internal/routes/widgets/weather"

	mediaSvc "server/internal/services/media"
	modulesSvc "server/internal/services/modules"
	promptsSvc "server/internal/services/prompts"
	caregiverSvc "server/internal/services/user/caregiver"
	newsSvc "server/internal/services/widgets/news"
	weatherSvc "server/internal/services/widgets/weather"

	_ "server/cmd/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func getEnvOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

// @title StartUp API
// @version 1.0
// @description This is the API documentation service
// @host localhost:8080
// @BasePath /
func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3001",
			"http://localhost:5173",
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// --- Database Setup ---
	dbHost := getEnvOrDefault("DB_HOST", "db")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("DB_USER", "myuser")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "mypassword")
	dbName := getEnvOrDefault("DB_NAME", "mydb")

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("failed to connect to DB:", err)
	}

	// --- MinIO Setup ---
	minioHost := getEnvOrDefault("MINIO_HOST", "minio")
	minioPort := getEnvOrDefault("MINIO_PORT", "9000")
	minioAccess := getEnvOrDefault("MINIO_ACCESS_KEY", "localuser")
	minioSecret := getEnvOrDefault("MINIO_SECRET_KEY", "localpass123")

	minioClient, err := minio.New(fmt.Sprintf("%s:%s", minioHost, minioPort), &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccess, minioSecret, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal("failed to connect to MinIO:", err)
	}

	// --- Ensure buckets exist ---
	ctx := context.Background()
	for _, bucket := range []string{"audio-bucket", "photo-bucket", "video-bucket"} {
		exists, err := minioClient.BucketExists(ctx, bucket)
		if err != nil {
			log.Fatalf("failed to check bucket %s: %v", bucket, err)
		}
		if !exists {
			err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				log.Fatalf("failed to create bucket %s: %v", bucket, err)
			}
			log.Println("created bucket:", bucket)
		} else {
			log.Println("bucket exists:", bucket)
		}
	}

	// --- Initialize Services ---
	audioService := mediaSvc.NewAudioService(db, minioClient, "audio-bucket")
	photoService := mediaSvc.NewPhotoService(db, minioClient, "photo-bucket")
	videoService := mediaSvc.NewVideoService(db, minioClient, "video-bucket")

	caregiverService := caregiverSvc.NewCaregiverService(db)
	newsService := newsSvc.NewNewsWidgetService(db, "api key")
	weatherService := weatherSvc.NewWeatherService(db, "api key")
	modulesService := modulesSvc.NewModuleService(db)

	eventsService := promptsSvc.NewEventService(db)
	messageService := promptsSvc.NewMessageService(db)
	caregiverLocationService := promptsSvc.NewCaregiverLocationService(db)
	carepartnerLocationService := promptsSvc.NewCarepartnerLocationService(db)

	// --- Controllers ---
	audioCtrl := mediaCtrlr.NewAudioController(audioService)
	photoCtrl := mediaCtrlr.NewPhotoController(photoService)
	videoCtrl := mediaCtrlr.NewVideoController(videoService)

	caregiverCtrl := caregiverCtrlr.NewCaregiverController(caregiverService)
	newsCtrl := newsCtrlr.NewNewsWidgetController(newsService)
	weatherCtrl := weatherCtrlr.NewWeatherWidgetController(weatherService)
	modulesCtrl := modulesCtrlr.NewModuleController(modulesService)

	eventsCtrl := promptsCtrlr.NewEventController(eventsService)
	messageCtrl := promptsCtrlr.NewMessageController(messageService)
	caregiverLocationCtrl := promptsCtrlr.NewCaregiverLocationController(caregiverLocationService)
	carepartnerLocationCtrl := promptsCtrlr.NewCarepartnerLocationController(carepartnerLocationService)

	// --- Routes ---
	mediaRts.RegisterAudioRoutes(r, audioCtrl)
	mediaRts.RegisterPhotoRoutes(r, photoCtrl)
	mediaRts.RegisterVideoRoutes(r, videoCtrl)

	caregiverRts.RegisterCaregiverRoutes(r, caregiverCtrl)
	newsRts.RegisterNewsWidgetRoutes(r, newsCtrl)
	weatherRts.RegisterWeatherWidgetRoutes(r, weatherCtrl)

	modulesRts.RegisterModulesRoutes(r, modulesCtrl)
	promptsRts.RegisterEventRoutes(r, eventsCtrl)
	promptsRts.RegisterMessageRoutes(r, messageCtrl)
	promptsRts.RegisterCaregiverLocationRoutes(r, caregiverLocationCtrl)
	promptsRts.RegisterCarepartnerLocationRoutes(r, carepartnerLocationCtrl)

	r.Mount("/swagger", httpSwagger.WrapHandler)

	// --- Start Server ---
	port := getEnvOrDefault("SERVER_PORT", "8080")
	fmt.Printf("Starting server on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
