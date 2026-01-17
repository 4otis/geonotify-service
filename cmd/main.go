package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/4otis/geonotify-service/config"
	"github.com/4otis/geonotify-service/internal/app"
)

// @title geonotify-service API
// @version 1.0
// @description REST API сервис для взаимодействия с
// @host localhost:8081
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}

	err = application.Run()
	if err != nil {
		log.Fatalf("error while running app: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	application.Stop()
}
