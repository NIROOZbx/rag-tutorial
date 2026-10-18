package handler

import (
	"net/http"

	"github.com/NIROOZbx/rag-tutorial.git/internal/document"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func UploadDocumentHandler(w http.ResponseWriter, r *http.Request) {
document.ChunkText("hello you are my sunshine",2)

}

func AskQuestionHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ask a question"))
}