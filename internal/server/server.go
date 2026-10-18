package server

import (
	
	"net/http"

	"github.com/NIROOZbx/rag-tutorial.git/internal/handler"
	"github.com/NIROOZbx/rag-tutorial.git/internal/worker"
)

func New(pool *worker.Pool) *http.ServeMux {
    mux := http.NewServeMux()

    mux.HandleFunc("GET /health", handler.HealthHandler)
    mux.HandleFunc("POST /question", handler.AskQuestionHandler)
    mux.HandleFunc("POST /upload", handler.UploadDocumentHandler(pool))

    return mux
}
