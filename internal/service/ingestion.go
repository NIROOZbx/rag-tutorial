package service

import (
	"context"

	"github.com/NIROOZbx/rag-tutorial.git/internal/document"
	"github.com/NIROOZbx/rag-tutorial.git/internal/embedding"
	"github.com/NIROOZbx/rag-tutorial.git/internal/repository"
	"github.com/NIROOZbx/rag-tutorial.git/internal/worker"
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

func (g *Ingestion) Ingest(ctx context.Context, fileName string, pdfData []byte) (int, error) {
	text, err := g.Pool.Submit(ctx, pdfData)
	if err != nil {
		return 0, err
	}

	chunks := document.ChunkText(text, 500)

	for i, chunk := range chunks {
		vector, err := g.GeminiClient.Embed(ctx, chunk)
		if err != nil {
			return 0, err
		}

		err = g.Repo.Save(ctx, repository.SaveParameters{
			DocumentName: fileName,
			ChunkIndex:   i,
			Content:      chunk,
			Embedding:    vector,
		})
		if err != nil {
			return 0, err
		}
	}

	return len(chunks), nil
}
