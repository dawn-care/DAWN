package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	authCtrlr "server/internal/controllers/auth"
	configurationCtrlr "server/internal/controllers/configurations"
	mediaCtrlr "server/internal/controllers/media"
	modulesCtrlr "server/internal/controllers/modules"
	promptsCtrlr "server/internal/controllers/prompts"
	caregiverCtrlr "server/internal/controllers/user/caregiver"
	newsCtrlr "server/internal/controllers/widgets/news"
	weatherCtrlr "server/internal/controllers/widgets/weather"

	configurationRts "server/internal/routes/configurations"
	mediaRts "server/internal/routes/media"
	modulesRts "server/internal/routes/modules"
	promptsRts "server/internal/routes/prompts"
	caregiverRts "server/internal/routes/user"
	newsRts "server/internal/routes/widgets/news"
	weatherRts "server/internal/routes/widgets/weather"

	authSvc "server/internal/services/auth"
	configurationSvc "server/internal/services/configurations"
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
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	httpSwagger "github.com/swaggo/http-swagger"

	mw "server/internal/middleware"
)

func getEnvOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3001",
			"http://localhost:5173",
			"http://localhost:8081",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// DB setup
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
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

	// MinIO setup
	minioHost := getEnvOrDefault("MINIO_HOST", "localhost")
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

	// Services
	audioService := mediaSvc.NewAudioService(db, minioClient, "audio-bucket")
	photoService := mediaSvc.NewPhotoService(db, minioClient, "photo-bucket")
	videoService := mediaSvc.NewVideoService(db, minioClient, "video-bucket")

	caregiverService := caregiverSvc.NewCaregiverService(db)
	newsService := newsSvc.NewNewsWidgetService(db, "3bc060a2d2c146bf92e6d9b94341d739")
	sportsService := newsSvc.NewSportsNewsWidgetService(db)
	weatherService := weatherSvc.NewWeatherService(db, "30d992579f0ac7f70e065cf0afaa070d")
	modulesService := modulesSvc.NewModuleService(db)
	configurationService := configurationSvc.NewConfigurationService(db)

	eventsService := promptsSvc.NewEventService(db)
	messageService := promptsSvc.NewMessageService(db)
	caregiverLocationService := promptsSvc.NewCaregiverLocationService(db)
	carepartnerLocationService := promptsSvc.NewCarepartnerLocationService(db)

	// Auth services
	googleService := authSvc.NewGoogleAuthService()
	sessionService := authSvc.NewSessionService(db)
	authService := authSvc.NewAuthService(googleService, caregiverService, sessionService)
	authController := authCtrlr.AuthController{AuthService: authService}

	// Auth routes
	r.Post("/api/auth/google", authController.GoogleLogin)
	r.Post("/api/auth/logout", authController.Logout)

	// Protected route (test route - update later w protected routes)
	r.Group(func(r chi.Router) {
		r.Use(mw.RequireAuth(sessionService))
		r.Get("/api/me", authController.MeHandler)
		r.Put("/api/me", authController.UpdateMeHandler)
	})

	// Controllers
	audioCtrl := mediaCtrlr.NewAudioController(audioService)
	photoCtrl := mediaCtrlr.NewPhotoController(photoService)
	videoCtrl := mediaCtrlr.NewVideoController(videoService)

	caregiverCtrl := caregiverCtrlr.NewCaregiverController(caregiverService)
	configurationCtrl := configurationCtrlr.NewConfigurationController(configurationService)
	newsCtrl := newsCtrlr.NewNewsWidgetController(newsService)
	sportsCtrl := newsCtrlr.NewSportsNewsWidgetController(sportsService)
	weatherCtrl := weatherCtrlr.NewWeatherWidgetController(weatherService)
	modulesCtrl := modulesCtrlr.NewModuleController(modulesService)

	eventsCtrl := promptsCtrlr.NewEventController(eventsService)
	messageCtrl := promptsCtrlr.NewMessageController(messageService)
	caregiverLocationCtrl := promptsCtrlr.NewCaregiverLocationController(caregiverLocationService)
	carepartnerLocationCtrl := promptsCtrlr.NewCarepartnerLocationController(carepartnerLocationService)

	// Routes
	mediaRts.RegisterAudioRoutes(r, audioCtrl)
	mediaRts.RegisterPhotoRoutes(r, photoCtrl)
	mediaRts.RegisterVideoRoutes(r, videoCtrl)

	caregiverRts.RegisterCaregiverRoutes(r, caregiverCtrl)
	configurationRts.RegisterConfigurationsRoutes(r, configurationCtrl)
	newsRts.RegisterNewsWidgetRoutes(r, newsCtrl)
	newsRts.RegisterSportsNewsWidgetRoutes(r, sportsCtrl)
	weatherRts.RegisterWeatherWidgetRoutes(r, weatherCtrl)

	modulesRts.RegisterModulesRoutes(r, modulesCtrl)
	promptsRts.RegisterEventRoutes(r, eventsCtrl)
	promptsRts.RegisterMessageRoutes(r, messageCtrl)
	promptsRts.RegisterCaregiverLocationRoutes(r, caregiverLocationCtrl)
	promptsRts.RegisterCarepartnerLocationRoutes(r, carepartnerLocationCtrl)

	r.Mount("/swagger", httpSwagger.WrapHandler)

	port := getEnvOrDefault("PORT", "8080")
	fmt.Printf("Starting server on port %s...\n", port)
	http.ListenAndServe(":"+port, r)
}
