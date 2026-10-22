package server

import (
	"net/http"

	"github.com/NIROOZbx/rag-tutorial.git/internal/handler"
)

func New(uploadHandler *handler.UploadHandler, questionHandler *handler.QuestionHandler) http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.HealthHandler)
	mux.HandleFunc("POST /question", questionHandler.AskQuestionHandler)
	mux.HandleFunc("POST /upload", uploadHandler.UploadDocumentHandler)

	return cors(mux)
}

// cors allows the deployed document reader frontend to call this API.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") == "https://doc-reader.nirooz.xyz" {
			w.Header().Set("Access-Control-Allow-Origin", "https://doc-reader.nirooz.xyz")
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
