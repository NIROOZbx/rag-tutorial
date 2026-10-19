package server

import (
	"net/http"

	"github.com/NIROOZbx/rag-tutorial.git/internal/handler"
)

func New(h *handler.UploadHandler) *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.HealthHandler)
	mux.HandleFunc("POST /question", handler.AskQuestionHandler)
	mux.HandleFunc("POST /upload", h.UploadDocumentHandler)

	return mux
}
