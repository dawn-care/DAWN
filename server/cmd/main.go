package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	authCtrlr "server/internal/controllers/auth"
	configurationCtrlr "server/internal/controllers/configurations"
	exceptionCtrlr "server/internal/controllers/exceptions"
	mediaCtrlr "server/internal/controllers/media"
	modulesCtrlr "server/internal/controllers/modules"
	playbackCtrlr "server/internal/controllers/playback"
	promptsCtrlr "server/internal/controllers/prompts"
	realtimeCtrlr "server/internal/controllers/realtime"
	scheduledItemsCtrlr "server/internal/controllers/scheduled_items"
	screenCtrlr "server/internal/controllers/screen"
	screenSettingsCtrlr "server/internal/controllers/screen_settings"
	screenWidgetsCtrlr "server/internal/controllers/screen_widgets"
	caregiverCtrlr "server/internal/controllers/user/caregiver"
	pairingCtrlr "server/internal/controllers/user/device_pairing"
	newsCtrlr "server/internal/controllers/widgets/news"
	weatherCtrlr "server/internal/controllers/widgets/weather"

	configurationRts "server/internal/routes/configurations"
	exceptionRts "server/internal/routes/exceptions"
	mediaRts "server/internal/routes/media"
	modulesRts "server/internal/routes/modules"
	playbackRts "server/internal/routes/playback"
	promptsRts "server/internal/routes/prompts"
	realtimeRts "server/internal/routes/realtime"
	scheduledItemsRts "server/internal/routes/scheduled_items"
	screenRts "server/internal/routes/screen"
	screenSettingsRts "server/internal/routes/screen_settings"
	screenWidgetsRts "server/internal/routes/screen_widgets"
	caregiverRts "server/internal/routes/user"
	pairingRts "server/internal/routes/user"
	newsRts "server/internal/routes/widgets/news"
	weatherRts "server/internal/routes/widgets/weather"

	authSvc "server/internal/services/auth"
	configurationSvc "server/internal/services/configurations"
	exceptionSvc "server/internal/services/exceptions"
	mediaSvc "server/internal/services/media"
	modulesSvc "server/internal/services/modules"
	playbackSvc "server/internal/services/playback"
	promptsSvc "server/internal/services/prompts"
	scheduledItemsSvc "server/internal/services/scheduled_items"
	screenSvc "server/internal/services/screen"
	screenSettingsSvc "server/internal/services/screen_settings"
	screenWidgetsSvc "server/internal/services/screen_widgets"
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
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/crypto/acme/autocert"

	database "server/internal/db"
	mw "server/internal/middleware"
	"server/internal/realtime"
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
			"http://localhost:5173",
			"http://localhost:8081",
			"http://localhost:8082", // staff app web build in development
			"http://startup-da.gccis.rit.edu",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// DB setup uses the same settings as database tooling.
	db, err := database.Open(context.Background())
	if err != nil {
		log.Fatal("failed to connect to DB: ", err)
	}
	defer db.Close()

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
	authService := authSvc.NewAuthService(googleService, caregiverService, carepartnerService, sessionService)

	// Agenda
	modulesService := modulesSvc.NewModuleService(db)
	configurationService := configurationSvc.NewConfigurationService(db)
	exceptionService := exceptionSvc.NewExceptionService(db)
	playbackService := playbackSvc.NewPlaybackService(db)

	// Screen
	screenSettingsService := screenSettingsSvc.NewScreenSettingsService(db)
	screenWidgetsService := screenWidgetsSvc.NewScreenWidgetService(db)
	scheduledItemsService := scheduledItemsSvc.NewScheduledItemService(db)
	// Realtime: one WebSocket per paired tablet, used to push staff commands.
	tabletHub := realtime.NewHub()
	go playbackService.RunExpirySweep(context.Background(), 5*time.Second)

	// Media
	audioService := mediaSvc.NewAudioService(db, minioClient, "media")
	photoService := mediaSvc.NewPhotoService(db, minioClient, "media")
	videoService := mediaSvc.NewVideoService(db, minioClient, "media")
	mediaService := mediaSvc.NewMediaService(db, minioClient, "media")

	// Widgets
	openWeatherKey := getEnvOrDefault("OPENWEATHER_API_KEY", "")
	newsAPIKey := getEnvOrDefault("NEWS_API_KEY", "")
	newsService := newsSvc.NewNewsWidgetService(db, newsAPIKey)
	sportsService := newsSvc.NewSportsNewsWidgetService(db)
	weatherService := weatherSvc.NewWeatherService(db, openWeatherKey)

	// Screen (the tablet's view; uses the widget and schedule services above)
	screenService := screenSvc.NewScreenService(db, weatherService, sportsService, newsService, scheduledItemsService)

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
	exceptionCtrl := exceptionCtrlr.NewExceptionController(exceptionService)
	playbackCtrl := playbackCtrlr.NewPlaybackController(playbackService, tabletHub)
	realtimeCtrl := realtimeCtrlr.NewRealtimeController(playbackService, carepartnerService, tabletHub, []string{
		// Browser origins allowed to open a tablet WebSocket (the tablet's web build
		// in development). The native Android app sends no Origin header.
		"localhost:8081",
		"startup-da.gccis.rit.edu",
	})

	// Screen
	screenSettingsCtrl := screenSettingsCtrlr.NewScreenSettingsController(screenSettingsService)
	screenWidgetsCtrl := screenWidgetsCtrlr.NewScreenWidgetController(screenWidgetsService)
	scheduledItemsCtrl := scheduledItemsCtrlr.NewScheduledItemController(scheduledItemsService)
	screenCtrl := screenCtrlr.NewScreenController(screenService)

	// Media
	audioCtrl := mediaCtrlr.NewAudioController(audioService)
	photoCtrl := mediaCtrlr.NewPhotoController(photoService)
	videoCtrl := mediaCtrlr.NewVideoController(videoService)
	mediaCtrl := mediaCtrlr.NewMediaController(mediaService)

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
	exceptionRts.RegisterExceptionRoutes(r, exceptionCtrl)
	playbackRts.RegisterPlaybackRoutes(r, playbackCtrl)
	realtimeRts.RegisterRealtimeRoutes(r, realtimeCtrl)

	// Screen
	screenSettingsRts.RegisterScreenSettingsRoutes(r, screenSettingsCtrl)
	screenWidgetsRts.RegisterScreenWidgetRoutes(r, screenWidgetsCtrl)
	scheduledItemsRts.RegisterScheduledItemRoutes(r, scheduledItemsCtrl)
	screenRts.RegisterScreenRoutes(r, screenCtrl)

	// Media
	mediaRts.RegisterAudioRoutes(r, audioCtrl)
	mediaRts.RegisterPhotoRoutes(r, photoCtrl)
	mediaRts.RegisterVideoRoutes(r, videoCtrl)
	mediaRts.RegisterMediaRoutes(r, mediaCtrl)

	// Widgets
	newsRts.RegisterNewsWidgetRoutes(r, newsCtrl)
	newsRts.RegisterSportsNewsWidgetRoutes(r, sportsCtrl)
	newsRts.RegisterSportsTeamsRoutes(r)
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

	// ===============================
	// SERVER STARTUP
	// ===============================
	runMode := getEnvOrDefault("RUN_MODE", "production")

	if runMode == "dev" {
		port := getEnvOrDefault("PORT", "8080")
		log.Printf("Starting HTTP dev server on :%s...\n", port)
		log.Fatal(http.ListenAndServe(":"+port, r))
	} else {
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
}
