package app

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/4otis/geonotify-service/config"
	_ "github.com/4otis/geonotify-service/docs"
	"github.com/4otis/geonotify-service/internal/adapter/repo/postgres"
	"github.com/4otis/geonotify-service/internal/cases"
	httphandler "github.com/4otis/geonotify-service/internal/handler/http"
	"github.com/4otis/geonotify-service/internal/worker"
	"github.com/4otis/geonotify-service/pkg/logger"
	"github.com/4otis/geonotify-service/pkg/redis"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type App struct {
	config        *config.Config
	logger        *zap.Logger
	httpServer    *http.Server
	dbPool        *pgxpool.Pool
	redisClient   *redis.Client
	webhookWorker *worker.WebhookWorker
}

func New(cfg *config.Config) (*App, error) {
	zapLogger, err := logger.NewDevelopment(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	app := &App{
		config: cfg,
		logger: zapLogger,
	}

	if err := app.initDB(); err != nil {
		return nil, err
	}

	if err := app.initRedis(); err != nil {
		return nil, err
	}

	if err := app.initUseCasesAndHandlers(); err != nil {
		return nil, err
	}

	if err := app.initWebhookWorker(); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) initDB() error {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, a.config.DBURL)
	if err != nil {
		return err
	}
	a.dbPool = pool

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	a.logger.Info("Database connected successfully")
	return nil
}

func (a *App) initRedis() error {
	ctx := context.Background()

	redisClient, err := redis.NewClient(ctx, a.config.RedisURL)
	if err != nil {
		return err
	}
	a.redisClient = redisClient

	a.logger.Info("Redis connected successfully")
	return nil
}

func (a *App) initWebhookWorker() error {
	webhookRepo := postgres.NewWebhookRepo(a.dbPool)

	a.webhookWorker = worker.NewWebhookWorker(
		a.logger,
		webhookRepo,
		a.redisClient,
		a.config.WebhookURL,
		a.config.MaxRetries,
		a.config.RetryDelaySeconds,
	)

	return nil
}

func (a *App) initUseCasesAndHandlers() error {
	incidentRepo := postgres.NewIncidentRepo(a.dbPool)
	checkRepo := postgres.NewCheckRepo(a.dbPool)
	webhookRepo := postgres.NewWebhookRepo(a.dbPool)

	locationUseCase := cases.NewLocationUseCase(
		incidentRepo,
		checkRepo,
		webhookRepo,
		a.redisClient,
		a.logger,
		a.config.CacheTTLMinutes,
	)
	incidentUseCase := cases.NewIncidentUseCase(
		incidentRepo,
		locationUseCase,
		a.logger,
	)
	statsUseCase := cases.NewStatsUseCase(
		incidentRepo,
		checkRepo,
		webhookRepo,
		a.logger,
	)

	httpIncidentHandler := httphandler.NewIncidentHandler(
		a.logger,
		incidentUseCase,
	)
	httpLocationHandler := httphandler.NewLocationHandler(
		a.logger,
		locationUseCase,
	)
	httpStatsHandler := httphandler.NewStatsHandler(
		a.logger,
		statsUseCase,
		a.config.StatsTimeWindowMinutes,
	)
	httpHealthHandler := httphandler.NewHealthHandler(
		a.logger,
		a.dbPool,
		a.redisClient,
		statsUseCase,
	)

	r := chi.NewRouter()

	r.Use(logger.Log(a.logger))
	r.Use(middleware.Timeout(30 * time.Second))

	r.Post("/api/v1/location/check", httpLocationHandler.LocationCheck)
	r.Get("/api/v1/incidents/stats", httpStatsHandler.GetStats)
	r.Get("/api/v1/system/health", httpHealthHandler.HealthCheck)

	r.Route("/api/v1/incidents", func(r chi.Router) {
		r.Use(a.apiKeyMiddleware)

		r.Post("/", httpIncidentHandler.IncidentCreate)
		r.Get("/", httpIncidentHandler.IncidentList)
		r.Get("/{incident_id}", httpIncidentHandler.IncidentGet)
		r.Put("/{incident_id}", httpIncidentHandler.IncidentUpdate)
		r.Delete("/{incident_id}", httpIncidentHandler.IncidentDelete)
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	a.httpServer = &http.Server{
		Addr:    ":" + a.config.HTTPPort,
		Handler: r,
	}

	return nil
}

func (a *App) apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			a.respondWithError(w, http.StatusUnauthorized, "authorization header is required")
			return
		}

		const bearerPrefix = "Bearer "
		if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			a.respondWithError(w, http.StatusUnauthorized, "authorization header must be in 'Bearer {token}' format")
			return
		}

		apiKey := authHeader[len(bearerPrefix):]
		if apiKey != a.config.APIKey {
			a.respondWithError(w, http.StatusUnauthorized, "invalid API key")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *App) respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errorResponse := map[string]interface{}{
		"error":   http.StatusText(code),
		"message": message,
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		a.logger.Error("failed to encode error response", zap.Error(err))
	}
}

func (a *App) Run() error {
	ctx := context.Background()
	a.webhookWorker.Start(ctx)

	go func() {
		a.logger.Info("Starting HTTP server",
			zap.String("port", a.config.HTTPPort),
			zap.String("env", a.config.LogLevel))

		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	return nil
}

func (a *App) Stop() {
	a.logger.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	if a.webhookWorker != nil {
		a.webhookWorker.Stop()
	}

	if a.dbPool != nil {
		a.dbPool.Close()
		a.logger.Info("Database connection closed")
	}

	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			a.logger.Error("Redis connection close error", zap.Error(err))
		}
		a.logger.Info("Redis connection closed")
	}

	a.logger.Sync()
	a.logger.Info("Servers stopped gracefully")
}
