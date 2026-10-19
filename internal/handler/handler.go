package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/NIROOZbx/rag-tutorial.git/internal/service"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"data": "Health check ok",
	}

	json.NewEncoder(w).Encode(resp)
}

type UploadHandler struct {
	IngestionService *service.Ingestion
}

func NewUploadDocumentHandler(svc *service.Ingestion) *UploadHandler {
	return &UploadHandler{
		IngestionService: svc,
	}
}

func (u *UploadHandler) UploadDocumentHandler(w http.ResponseWriter, r *http.Request) {

	file, header, err := r.FormFile("document")
	if err != nil {
		http.Error(w, "Missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	pdfName := header.Filename

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}

	chunks, err := u.IngestionService.Ingest(r.Context(), pdfName, data)
	if err != nil {
		http.Error(w, "PDF processing failed",
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "PDF processed successfully",
		"chunks":  chunks,
	})

}

func AskQuestionHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ask a question"))
}
