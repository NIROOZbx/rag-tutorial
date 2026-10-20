package service

import (
	"context"
	"fmt"
	"strings"
	"github.com/google/uuid"

	"github.com/NIROOZbx/rag-tutorial.git/internal/embedding"
	"github.com/NIROOZbx/rag-tutorial.git/internal/repository"
)

type AnswerService struct {
	embedder  *embedding.GeminiClient
	generator AnswerGenerator
	ChunkRepo *repository.ChunkRepository
}

type AnswerGenerator interface {
	GenerateAnswer(ctx context.Context, prompt string) (string, error)
}

func NewAnswerService(repo *repository.ChunkRepository, embedder *embedding.GeminiClient, generator AnswerGenerator) *AnswerService {
	return &AnswerService{
		embedder:  embedder,
		generator: generator,
		ChunkRepo: repo,
	}
}

func (a *AnswerService) RetrieveAnswer(ctx context.Context, question string,docID uuid.UUID) (string, error) {

	vector, err := a.embedder.Embed(ctx, question)
	if err != nil {
		return "", err
	}

	results, err := a.ChunkRepo.Search(ctx, vector,docID)
	if err != nil {
		return "", err
	}

	 if len(results) == 0 {
        return "I couldn't find any chunks for this document.", nil
    }

	var contextBuilder strings.Builder

	for _, result := range results {
		    contextBuilder.WriteString("PDF: ")
    contextBuilder.WriteString(result.DocumentName)
    contextBuilder.WriteString("\n")
    contextBuilder.WriteString(result.Content)
    contextBuilder.WriteString("\n\n")
	}

	contextText := contextBuilder.String()


	prompt := fmt.Sprintf(`Answer the question using only the context below 
	If the answer isn't in the context, say you don't know.
	Context:%s
	Question:%s`, 
	contextText, question)

	return a.generator.GenerateAnswer(ctx, prompt)

}
