package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NIROOZbx/rag-tutorial.git/internal/embedding"
	"github.com/NIROOZbx/rag-tutorial.git/internal/server"
	"github.com/NIROOZbx/rag-tutorial.git/internal/worker"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("load .env file: %v", err)
	}

	ctx := context.Background()
	if _, err := embedding.NewGeminiClient(ctx); err != nil {
		log.Fatalf("initialize Gemini client: %v", err)
	}

	pdfPool, err := worker.NewWorkerPool(20, 3)
	if err != nil {
		log.Fatalf("create PDF worker pool: %v", err)
	}
	defer pdfPool.Close()

	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           server.New(pdfPool),
		ReadHeaderTimeout: 5 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	go func() {
		<-stop
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server: %v", err)
	}
}
