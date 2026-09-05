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
)

func main() {
	// load config
	cfg := config.MustLoad()
	// database setup

	//setup router
	router := http.NewServeMux()
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Welcomme to students api"))
	})

	//setup server

	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}
	slog.Info("Server Started ",slog.String("address",cfg.Addr))
	fmt.Printf("Server started %s", cfg.HTTPServer.Addr)
	done := make(chan os.Signal,1)
	signal.Notify(done,os.Interrupt,syscall.SIGINT,syscall.SIGTERM)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("Failed to start server")
		}

	}()
// gacefull shut down
	<-done
	slog.Info("shutting down the server")
	ctx,cancel :=	context.WithTimeout(context.Background(),5 * time.Second)

	defer cancel()


	err:=server.Shutdown(ctx)

	if err != nil{
		slog.Error("failed to shutdown server", slog.String("error",err.Error()))

	}
	slog.Info("Server shutdown successfully")


}
