package app

import (
	"context"
	"net/http"
	"time"

	"github.com/4otis/geonotify-service/config"

	_ "github.com/4otis/geonotify-service/docs"
	"github.com/4otis/geonotify-service/internal/adapter/repo/postgres"
	"github.com/4otis/geonotify-service/internal/cases"
	httphandler "github.com/4otis/geonotify-service/internal/handler/http"
	"github.com/4otis/geonotify-service/pkg/logger"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type App struct {
	config     *config.Config
	logger     *zap.Logger
	httpServer *http.Server
	dbPool     *pgxpool.Pool
}

func New(cfg *config.Config) (*App, error) {
	zapLogger, err := logger.NewDevelopment("")
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

	if err := app.initServers(); err != nil {
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

func (a *App) initServers() error {
	incidentRepo := postgres.NewIncidentRepo(a.dbPool)
	incidentUseCase := cases.NewIncidentUseCase(incidentRepo)

	httpIncidentHandler := httphandler.NewIncidentHandler(
		a.logger,
		"1",
		incidentUseCase,
	)
	r := chi.NewRouter()
	r.Use(logger.Log(a.logger))
	r.Post("/api/v1/incidents", httpIncidentHandler.IncidentCreate)
	r.Get("/api/v1/incidents/{incident_id}", httpIncidentHandler.IncidentGet)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	a.httpServer = &http.Server{
		Addr:    ":" + a.config.HTTPPort,
		Handler: r,
	}

	return nil
}

func (a *App) Run() error {
	go func() {
		a.logger.Info("Starting HTTP server", zap.String("port", a.config.HTTPPort))
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	return nil
}

func (a *App) Stop() {
	a.logger.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	if a.dbPool != nil {
		a.dbPool.Close()
	}

	a.logger.Sync()
	a.logger.Info("Servers stopped gracefully")
}
