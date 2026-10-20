package server

import (
	"net/http"

	"github.com/NIROOZbx/rag-tutorial.git/internal/handler"
)

func New(uploadHandler *handler.UploadHandler, questionHandler *handler.QuestionHandler) *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.HealthHandler)
	mux.HandleFunc("POST /question", questionHandler.AskQuestionHandler)
	mux.HandleFunc("POST /upload", uploadHandler.UploadDocumentHandler)

	return mux
}
