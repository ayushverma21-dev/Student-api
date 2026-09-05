package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayushverma21-dev/Student-api/internal/config"
	"github.com/ayushverma21-dev/Student-api/internal/http/handler/student"
	"github.com/ayushverma21-dev/Student-api/internal/storage/postgres"
)

func main() {
	// load config
	cfg := config.MustLoad()
	// database setup
	storage, err := postgres.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("storage initialized", slog.String("env",cfg.Env),slog.String("version","1.0.0"))
	//setup router
	router := http.NewServeMux()


	router.HandleFunc("POST /api/students", student.New(storage))
	router.HandleFunc("GET /api/students/{id}",student.GetById(storage))
	router.HandleFunc("GET /api/students", student.GetList(storage))

	//setup server

	server := http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
	}
	slog.Info("Server Started ", slog.String("address", cfg.HTTPServer.Addr))
	fmt.Printf("Server started %s", cfg.HTTPServer.Addr)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("Failed to start server")
		}

	}()
	// gacefull shut down
	<-done
	slog.Info("shutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = server.Shutdown(ctx)

	if err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))

	}
	slog.Info("Server shutdown successfully")

}
