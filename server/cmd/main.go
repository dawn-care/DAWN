package main

import (
	"fmt"
	"os"

	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"database/sql"
	"log"
	"net/http"

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
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// @title StartUp API
// @version 1.0
// @description This is the API documentation service
// @host localhost:8080
// @BasePath /
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
	dbHost := os.Getenv("DB_HOST")         // "db"
	dbPort := os.Getenv("DB_PORT")         // "5432"
	dbUser := os.Getenv("DB_USER")         // "myuser"
	dbPassword := os.Getenv("DB_PASSWORD") // "mypassword"
	dbName := os.Getenv("DB_NAME")         // "mydb"

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, dberr := sql.Open("postgres", connStr)
	if dberr != nil {
		log.Fatal("failed to connect to DB:", dberr)
	}

	// MINIO SETUP
	minioClient, bloberr := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minio-access-key", "minio-secret-key", ""),
		Secure: false,
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
	// Widgets
	newsService := newsSvc.NewNewsWidgetService(db, "api key (insert here)")
	// sportsNewsService := newsSvc.NewSportsNewsWidgetService(db) // see TODO in ctrlrs initialization
	weatherService := weatherSvc.NewWeatherService(db, "api key (insert here)")
	// Modules
	modulesService := modulesSvc.NewModuleService(db)
	// Prompts
	eventsService := promptsSvc.NewEventService(db)
	messageService := promptsSvc.NewMessageService(db)
	caregiverLocationService := promptsSvc.NewCaregiverLocationService(db)
	carepartnerLocationService := promptsSvc.NewCarepartnerLocationService(db)

	// INITIALIZE CTRLRS
	// Media
	audioCtrl := mediaCtrlr.NewAudioController(audioService)
	photoCtrl := mediaCtrlr.NewPhotoController(photoService)
	videoCtrl := mediaCtrlr.NewVideoController(videoService)
	// User
	caregiverCtrl := caregiverCtrlr.NewCaregiverController(caregiverService)
	// Widgets
	newsCtrl := newsCtrlr.NewNewsWidgetController(newsService)
	// sportsNewsCtrl := newsCtrlr.NewSportsNewsWidgetController(sportsNewsService) // TODO: mismatch between service and controller, doesn't properly match interface
	weatherCtrl := weatherCtrlr.NewWeatherWidgetController(weatherService)
	// Modules
	modulesCtrl := modulesCtrlr.NewModuleController(modulesService)
	// Prompts
	eventsCtrl := promptsCtrlr.NewEventController(eventsService)
	messageCtrl := promptsCtrlr.NewMessageController(messageService)
	caregiverLocationCtrl := promptsCtrlr.NewCaregiverLocationController(caregiverLocationService)
	carepartnerLocationCtrl := promptsCtrlr.NewCarepartnerLocationController(carepartnerLocationService)

	// ROUTE INITIALIZATION
	// Media
	mediaRts.RegisterAudioRoutes(r, audioCtrl)
	mediaRts.RegisterPhotoRoutes(r, photoCtrl)
	mediaRts.RegisterVideoRoutes(r, videoCtrl)
	// User
	caregiverRts.RegisterCaregiverRoutes(r, caregiverCtrl)
	// Widgets
	newsRts.RegisterNewsWidgetRoutes(r, newsCtrl)
	// newsRts.RegisterSportsNewsWidgetRoutes(r, sportsNewsCtrl)
	weatherRts.RegisterWeatherWidgetRoutes(r, weatherCtrl)
	// Modules
	modulesRts.RegisterModulesRoutes(r, modulesCtrl)
	// Prompts
	promptsRts.RegisterEventRoutes(r, eventsCtrl)
	promptsRts.RegisterMessageRoutes(r, messageCtrl)
	promptsRts.RegisterCaregiverLocationRoutes(r, caregiverLocationCtrl)
	promptsRts.RegisterCarepartnerLocationRoutes(r, carepartnerLocationCtrl)

	// Swaggo Docs
	r.Mount("/swagger", httpSwagger.WrapHandler)

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
