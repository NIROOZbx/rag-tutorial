package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/NIROOZbx/rag-tutorial.git/internal/service"
	"github.com/google/uuid"
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

	docID, chunks, err := u.IngestionService.Ingest(
		r.Context(),
		pdfName,
		data,
	)
	if err != nil {
		log.Println("PDF processing failed:", err)
		http.Error(w, "PDF processing failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message":       "PDF processed successfully",
		"document_id":   docID,
		"document_name": pdfName,
		"chunks":        chunks,
	})

}

type QuestionHandler struct {
	AnswerService *service.AnswerService
}

func NewQuestionHandler(answerService *service.AnswerService) *QuestionHandler {
	return &QuestionHandler{AnswerService: answerService}
}

func (h *QuestionHandler) AskQuestionHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Question string    `json:"question"`
		DocID    uuid.UUID `json:"doc_id"`
	}

	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(request.Question) == "" {
		http.Error(w, "Question is required", http.StatusBadRequest)
		return
	}

	if request.DocID == uuid.Nil {
    http.Error(w, "Document ID is required", http.StatusBadRequest)
    return
}

	answer, err := h.AnswerService.RetrieveAnswer(r.Context(), request.Question, request.DocID)
	if err != nil {
		http.Error(w, "Failed to answer question", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"answer": answer}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
