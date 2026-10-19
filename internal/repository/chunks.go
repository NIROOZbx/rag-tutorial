
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type ChunkRepository struct {
	db *pgxpool.Pool
}

func NewChunkRepository(db *pgxpool.Pool) *ChunkRepository {
	return &ChunkRepository{db: db}
}

type SaveParameters struct {
	DocumentName string
	ChunkIndex   int
	Content      string
	Embedding    []float32
}

func (r *ChunkRepository) Save(ctx context.Context,args SaveParameters) error {
	if len(args.Embedding) != 3072 {
		return fmt.Errorf(
			"expected 3072 embedding dimensions, got %d",
			len(args.Embedding),
		)
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO document_chunks
			(document_name, chunk_index, content, embedding)
		VALUES ($1, $2, $3, $4)
	`,
		args.DocumentName,
		args.ChunkIndex,
		args.Content,
		pgvector.NewVector(args.Embedding),
	)
	if err != nil {
		return fmt.Errorf("save chunk: %w", err)
	}

	return nil
}
