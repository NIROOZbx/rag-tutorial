package service

import (
	"context"

	"github.com/NIROOZbx/rag-tutorial.git/internal/document"
	"github.com/NIROOZbx/rag-tutorial.git/internal/embedding"
	"github.com/NIROOZbx/rag-tutorial.git/internal/repository"
	"github.com/NIROOZbx/rag-tutorial.git/internal/worker"
	"github.com/google/uuid"
)

type Ingestion struct {
	Pool         *worker.Pool
	GeminiClient *embedding.GeminiClient
	Repo         *repository.ChunkRepository
}

func NewIngestionService(pool *worker.Pool, geminiClient *embedding.GeminiClient, Repo *repository.ChunkRepository) *Ingestion {
	return &Ingestion{
		Pool:         pool,
		GeminiClient: geminiClient,
		Repo:         Repo,
	}
}

func (g *Ingestion) Ingest(ctx context.Context, fileName string, pdfData []byte) (uuid.UUID,int, error) {
	text, err := g.Pool.Submit(ctx, pdfData)
    if err != nil {
        return uuid.Nil, 0, err
    }

    chunks := document.ChunkText(text, 500)

    docID, err := g.Repo.CreateDocument(ctx, fileName)
    if err != nil {
        return uuid.Nil, 0, err
    }

    for i, chunk := range chunks {
        vector, err := g.GeminiClient.Embed(ctx, chunk)
        if err != nil {
            return docID, 0, err
        }

        err = g.Repo.Save(ctx, repository.SaveParameters{
            DocID:      docID,
            ChunkIndex: i,
            Content:    chunk,
            Embedding:  vector,
        })
        if err != nil {
            return docID, 0, err
        }
    }

    return docID, len(chunks), nil
}


