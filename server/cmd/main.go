package main

import (
	"context"
	"crypto/tls"
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
	pairingCtrlr "server/internal/controllers/user/device_pairing"
	newsCtrlr "server/internal/controllers/widgets/news"
	weatherCtrlr "server/internal/controllers/widgets/weather"

	configurationRts "server/internal/routes/configurations"
	mediaRts "server/internal/routes/media"
	modulesRts "server/internal/routes/modules"
	promptsRts "server/internal/routes/prompts"
	caregiverRts "server/internal/routes/user"
	pairingRts "server/internal/routes/user"
	newsRts "server/internal/routes/widgets/news"
	weatherRts "server/internal/routes/widgets/weather"

	authSvc "server/internal/services/auth"
	configurationSvc "server/internal/services/configurations"
	mediaSvc "server/internal/services/media"
	modulesSvc "server/internal/services/modules"
	promptsSvc "server/internal/services/prompts"
	caregiverSvc "server/internal/services/user/caregiver"
	pairingSvc "server/internal/services/user/device_pairing"
	newsSvc "server/internal/services/widgets/news"
	weatherSvc "server/internal/services/widgets/weather"

	_ "server/cmd/docs"
	carepartnerCtrlr "server/internal/controllers/user/carepartner"
	carepartnerRts "server/internal/routes/user"
	carepartnerSvc "server/internal/services/user/carepartner"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/crypto/acme/autocert"

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

	// url := getEnvOrDefault("API_URL", "localhost")

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins: []string{
		// 	fmt.Sprintf("%s:3001", url),
		// 	fmt.Sprintf("%s:5173", url),
		// 	fmt.Sprintf("%s:8081", url),
		// 	fmt.Sprintf("%s", url),
		// 	"http://startup-da.gccis.rit.edu",
		// }
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
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

	// --- Ensure bucket exists ---
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, "media")
	if err != nil {
		log.Fatalf("failed to check bucket media: %v", err)
	}
	if !exists {
		err := minioClient.MakeBucket(ctx, "media", minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("failed to create bucket media: %v", err)
		}
		log.Println("created bucket: media")
	} else {
		log.Println("bucket exists: media")
	}

	// ------------ SERVICES ------------
	// User
	caregiverService := caregiverSvc.NewCaregiverService(db)
	devicePairingService := pairingSvc.NewDevicePairingService(db)
	carepartnerService := carepartnerSvc.NewCarepartnerService(db)

	// Auth
	googleService := authSvc.NewGoogleAuthService()
	sessionService := authSvc.NewSessionService(db)
	authService := authSvc.NewAuthService(googleService, caregiverService, sessionService)

	// Agenda
	modulesService := modulesSvc.NewModuleService(db)
	configurationService := configurationSvc.NewConfigurationService(db)

	// Media
	audioService := mediaSvc.NewAudioService(db, minioClient, "media")
	photoService := mediaSvc.NewPhotoService(db, minioClient, "media")
	videoService := mediaSvc.NewVideoService(db, minioClient, "media")

	// Widgets
	newsService := newsSvc.NewNewsWidgetService(db, "api key")
	sportsService := newsSvc.NewSportsNewsWidgetService(db)
	weatherService := weatherSvc.NewWeatherService(db, "api key")

	// Prompts
	messageService := promptsSvc.NewMessageService(db)
	eventsService := promptsSvc.NewEventService(db)
	caregiverLocationService := promptsSvc.NewCaregiverLocationService(db)
	carepartnerLocationService := promptsSvc.NewCarepartnerLocationService(db)

	// ------------ CONTROLLERS ------------
	// User
	caregiverCtrl := caregiverCtrlr.NewCaregiverController(caregiverService)
	devicePairingCtrl := pairingCtrlr.NewDevicePairingController(devicePairingService)
	carepartnerCtrl := carepartnerCtrlr.NewCarepartnerController(carepartnerService)

	// Auth
	authController := authCtrlr.AuthController{AuthService: authService}

	// Agenda
	modulesCtrl := modulesCtrlr.NewModuleController(modulesService)
	configurationCtrl := configurationCtrlr.NewConfigurationController(configurationService)

	// Media
	audioCtrl := mediaCtrlr.NewAudioController(audioService)
	photoCtrl := mediaCtrlr.NewPhotoController(photoService)
	videoCtrl := mediaCtrlr.NewVideoController(videoService)

	// Widgets
	newsCtrl := newsCtrlr.NewNewsWidgetController(newsService)
	sportsCtrl := newsCtrlr.NewSportsNewsWidgetController(sportsService)
	weatherCtrl := weatherCtrlr.NewWeatherWidgetController(weatherService)

	// Prompts
	messageCtrl := promptsCtrlr.NewMessageController(messageService)
	eventsCtrl := promptsCtrlr.NewEventController(eventsService)
	caregiverLocationCtrl := promptsCtrlr.NewCaregiverLocationController(caregiverLocationService)
	carepartnerLocationCtrl := promptsCtrlr.NewCarepartnerLocationController(carepartnerLocationService)

	// ------------ ROUTES ------------
	// User
	caregiverRts.RegisterCaregiverRoutes(r, caregiverCtrl)
	r.Route("/api/users", func(r chi.Router) {
		r.Post("/carepartner/create-carepartner", carepartnerCtrl.CreateCarepartner)
		r.Get("/carepartner/{id}", carepartnerCtrl.GetCarepartnerByID)
		r.Get("/carepartner", carepartnerCtrl.GetAllCarepartners)
		r.Put("/carepartner/update", carepartnerCtrl.UpdateCarepartner)
		r.Delete("/carepartner/{id}", carepartnerCtrl.DeleteCarepartner)

		r.Post("/device-pairing/code", devicePairingCtrl.GeneratePairingCode)
		r.Post("/device-pairing/verify", devicePairingCtrl.VerifyPairingCode)
		r.Post("/device-pairing/status", devicePairingCtrl.IsDevicePaired)

		pairingRts.RegisterDevicePairingRoutes(r, devicePairingCtrl)
		carepartnerRts.RegisterCarepartnerRoutes(r, carepartnerCtrl)
	})

	// Auth
	r.Post("/api/auth/google", authController.GoogleLogin)
	r.Post("/api/auth/logout", authController.Logout)

	// Protected route (test route - update later w protected routes)
	r.Group(func(r chi.Router) {
		r.Use(mw.RequireAuth(sessionService))
		r.Get("/api/me", authController.MeHandler)
		r.Put("/api/me", authController.UpdateMeHandler)
	})

	// Agenda
	modulesRts.RegisterModulesRoutes(r, modulesCtrl)
	configurationRts.RegisterConfigurationsRoutes(r, configurationCtrl)

	// Media
	mediaRts.RegisterAudioRoutes(r, audioCtrl)
	mediaRts.RegisterPhotoRoutes(r, photoCtrl)
	mediaRts.RegisterVideoRoutes(r, videoCtrl)

	// Widgets
	newsRts.RegisterNewsWidgetRoutes(r, newsCtrl)
	newsRts.RegisterSportsNewsWidgetRoutes(r, sportsCtrl)
	weatherRts.RegisterWeatherWidgetRoutes(r, weatherCtrl)

	// Prompts
	promptsRts.RegisterMessageRoutes(r, messageCtrl)
	promptsRts.RegisterEventRoutes(r, eventsCtrl)
	promptsRts.RegisterCaregiverLocationRoutes(r, caregiverLocationCtrl)
	promptsRts.RegisterCarepartnerLocationRoutes(r, carepartnerLocationCtrl)

	r.Mount("/swagger", httpSwagger.WrapHandler)
	// ===============================
	// FRONTEND (STATIC FILE SERVING)
	// ===============================

	// Serve static assets (JS, CSS, images)
	// FRONTEND (STATIC FILE SERVING)
	// Serve everything under /assets/* from /app/frontend/assets
	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("/app/frontend/assets"))))

	// Catch-all for SPA (any other path serves index.html)
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/app/frontend/index.html")
	})

	// port := getEnvOrDefault("PORT", "8080")
	// fmt.Printf("Starting server on port %s...\n", port)
	// http.ListenAndServe(":"+port, r)

	// TEMP
	// manager := &autocert.Manager{
	// 	Cache:      autocert.DirCache("/certs"),
	// 	Prompt:     autocert.AcceptTOS,
	// 	HostPolicy: autocert.HostWhitelist("startup-da.gccis.rit.edu"),
	// 	Client: &acme.Client{
	// 		DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
	// 	},
	// }

	// ACTUAL
	manager := &autocert.Manager{
		Cache:      autocert.DirCache("/certs"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("startup-da.gccis.rit.edu"),
	}

	server := &http.Server{
		Addr:    ":443",
		Handler: r,
		TLSConfig: &tls.Config{
			GetCertificate: manager.GetCertificate,
		},
	}

	go http.ListenAndServe(":80", manager.HTTPHandler(r))

	log.Println("Starting HTTPS server on 443")
	log.Fatal(server.ListenAndServeTLS("", ""))
}
