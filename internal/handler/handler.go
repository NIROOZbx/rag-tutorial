package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/NIROOZbx/rag-tutorial.git/internal/document"
	"github.com/NIROOZbx/rag-tutorial.git/internal/worker"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	resp:=map[string]string{
		"data":"Health check ok",
	}

	json.NewEncoder(w).Encode(resp)
}

func UploadDocumentHandler(pool *worker.Pool) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("document")
		if err != nil {
			http.Error(w, "Missing file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusBadRequest)
			return
		}

		text, err := pool.Submit(r.Context(), data)
		if err != nil {
			http.Error(w, "PDF processing failed", http.StatusInternalServerError)
			return
		}
		
		chunks := document.ChunkText(text, 500)

		w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]any{
            "message": "PDF processed successfully",
            "chunks":  len(chunks),
        })
	}
}

func AskQuestionHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ask a question"))
}
