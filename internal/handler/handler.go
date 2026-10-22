package handler

import (
	"encoding/json"
	"errors"
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

	const maxFileSize = 10 << 20 // 10 MB
	log.Printf("upload request received: method=%s content_type=%q content_length=%d origin=%q", r.Method, r.Header.Get("Content-Type"), r.ContentLength, r.Header.Get("Origin"))

	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	file, header, err := r.FormFile("document")
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			log.Printf("upload rejected: request exceeded %d byte limit", maxFileSize)
			http.Error(w, "File size exceeds 10 MB limit",
				http.StatusRequestEntityTooLarge)
			return
		}

		log.Printf("upload rejected: parse multipart form: %v", err)
		http.Error(w, "Invalid or missing file",
			http.StatusBadRequest)
		return
	}
	defer file.Close()

	pdfName := header.Filename

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("upload rejected: read uploaded file %q: %v", pdfName, err)
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}

	docID, chunks, err := u.IngestionService.Ingest(
		r.Context(),
		pdfName,
		data,
	)
	if err != nil {
		log.Printf("PDF processing failed: file=%q bytes=%d error=%+v", pdfName, len(data), err)
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
