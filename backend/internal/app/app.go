package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"football-api/internal/config"
	"football-api/internal/middleware"
	adminmod "football-api/internal/modules/admin"
	authmod "football-api/internal/modules/auth"
	matchesmod "football-api/internal/modules/matches"
	playersmod "football-api/internal/modules/players"
	teamsmod "football-api/internal/modules/teams"
	"football-api/pkg/storage"
)

func New(cfg config.Config, log *slog.Logger, db *pgxpool.Pool) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	var storageProvider storage.Storage
	if cldURL := cfg.CloudinaryURL(); cldURL != "" {
		if cldStorage, err := storage.NewCloudinaryStorage(cldURL); err == nil {
			storageProvider = cldStorage
		}
	}

	adminRepo := adminmod.NewRepository(db)
	authRepo := authmod.NewRepository(db)
	authSvc := authmod.NewService(adminRepo, authRepo, cfg.JWTSecret)
	authHandler := authmod.NewHandler(authSvc, log)
	authMiddleware := middleware.Auth(cfg.JWTSecret, authRepo, log)

	v1 := r.Group("/api/v1")
	authHandler.RegisterRoutes(v1.Group("/auth"), authMiddleware)

	protected := v1.Group("")
	protected.Use(authMiddleware)

	teamsRepo := teamsmod.NewRepository(db)
	teamsSvc := teamsmod.NewService(teamsRepo)
	teamsHandler := teamsmod.NewHandler(teamsSvc, storageProvider, log)
	teamsHandler.RegisterRoutes(protected.Group("/teams"))

	playersRepo := playersmod.NewRepository(db)
	playersSvc := playersmod.NewService(playersRepo, teamsRepo)
	playersHandler := playersmod.NewHandler(playersSvc, log)
	playersHandler.RegisterRoutes(protected.Group("/players"))

	matchesRepo := matchesmod.NewRepository(db)
	matchesSvc := matchesmod.NewService(matchesRepo, teamsRepo, playersRepo)
	matchesHandler := matchesmod.NewHandler(matchesSvc, log)
	matchesHandler.RegisterRoutes(protected.Group("/matches"))

	return &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
