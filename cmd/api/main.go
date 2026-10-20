package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NIROOZbx/rag-tutorial.git/internal/database"
	"github.com/NIROOZbx/rag-tutorial.git/internal/embedding"
	"github.com/NIROOZbx/rag-tutorial.git/internal/groq"
	"github.com/NIROOZbx/rag-tutorial.git/internal/handler"
	"github.com/NIROOZbx/rag-tutorial.git/internal/repository"
	"github.com/NIROOZbx/rag-tutorial.git/internal/server"
	"github.com/NIROOZbx/rag-tutorial.git/internal/service"
	"github.com/NIROOZbx/rag-tutorial.git/internal/worker"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("load .env file: %v", err)
	} else if err == nil {
		log.Print("loaded environment variables from .env")
	} else {
		log.Print(".env file not found; using process environment")
	}

	ctx := context.Background()

	geminiClient, err := embedding.NewGeminiClient(ctx)

	if err != nil {
		log.Fatalf("initialize Gemini client: %v", err)
	}
	log.Print("Gemini client initialized")

	answerGenerator, err := groq.NewClient(os.Getenv("GROQ_API_KEY"))
	if err != nil {
		log.Fatalf("initialize Groq client: %v", err)
	}
	log.Print("Groq answer generator initialized")

	databasePool, err := database.Connect(ctx, os.Getenv("SUPABASE_DB_URL"))
	if err != nil {
		log.Fatalf("connect to Supabase Postgres: %v", err)
	}
	log.Print("Supabase Postgres connected and ping succeeded")
	defer func() {
		databasePool.Close()
		log.Print("Supabase Postgres connection pool closed")
	}()

	pdfPool, err := worker.NewWorkerPool(20, 3)
	if err != nil {
		log.Fatalf("create PDF worker pool: %v", err)
	}
	log.Print("PDF worker pool created (workers=3, queue=20)")
	defer func() {
		pdfPool.Close()
		log.Print("PDF worker pool closed")
	}()

	chunkRepo := repository.NewChunkRepository(databasePool)

	ingestionService := service.NewIngestionService(pdfPool, geminiClient, chunkRepo)

	uploadHandler := handler.NewUploadDocumentHandler(ingestionService)
	answerService := service.NewAnswerService(chunkRepo, geminiClient, answerGenerator)
	questionHandler := handler.NewQuestionHandler(answerService)

	log.Print("ingestion service and HTTP handlers initialized")

	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           server.New(uploadHandler, questionHandler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := runServer(httpServer); err != nil {
		log.Printf("HTTP server: %v", err)
	}

}
func runServer(httpServer *http.Server) error {

	listener, err := net.Listen("tcp", httpServer.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", httpServer.Addr, err)
	}
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}
	}()

	log.Printf("HTTP server listening on %s", httpServer.Addr)

	err = httpServer.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}
