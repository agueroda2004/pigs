package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"server/internal/platform/config"
	appcontainer "server/internal/platform/container"
	platformhttp "server/internal/platform/http"
)

func main() {
	applicationConfig, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dependencies, err := appcontainer.New(ctx, applicationConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer dependencies.Close()

	mux := http.NewServeMux()
	userHandler := dependencies.User.Handler
	userHandler.RegisterRoutes(mux)

	authHandler := dependencies.Auth.Handler
	authHandler.RegisterRoutes(mux)

	breedHandler := dependencies.Breed.Handler
	breedHandler.RegisterRoutes(mux)

	boarHandler := dependencies.Boar.Handler
	boarHandler.RegisterRoutes(mux)

	sowHandler := dependencies.Sow.Handler
	sowHandler.RegisterRoutes(mux)

	operatorHandler := dependencies.Operator.Handler
	operatorHandler.RegisterRoutes(mux)

	serviceHandler := dependencies.Service.Handler
	serviceHandler.RegisterRoutes(mux)

	abortionHandler := dependencies.Abortion.Handler
	abortionHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    ":" + applicationConfig.Port,
		Handler: platformhttp.CORS(applicationConfig.CorsAllowedOrigins)(mux),
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatalf("No se pudo iniciar el servidor en %s: %v", server.Addr, err)
	}
	log.Printf("Servidor ejecutándose en http://localhost%s", server.Addr)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Serve(listener)
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("No se pudo detener el servidor correctamente: %v", err)
		}
	}
}
