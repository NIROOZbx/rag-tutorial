package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type ChunkRepository struct {
	db *pgxpool.Pool
}

func NewChunkRepository(db *pgxpool.Pool) *ChunkRepository {
	return &ChunkRepository{db: db}
}

type ChunkParameters struct {
	ChunkIndex int
	Content    string
	Embedding  []float32
}

type SearchResult struct {
	DocumentName string
	Content      string
}

func (r *ChunkRepository) SaveDocumentWithChunks(ctx context.Context, fileName string, chunks []ChunkParameters) (uuid.UUID, error) {
	for _, chunk := range chunks {
		if len(chunk.Embedding) != 3072 {
			return uuid.Nil, fmt.Errorf("chunk %d: expected 3072 embedding dimensions, got %d", chunk.ChunkIndex, len(chunk.Embedding))
		}
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin document transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	var docID uuid.UUID
	err = tx.QueryRow(ctx, `
        INSERT INTO documents (document_name)
        VALUES ($1)
        RETURNING id
    `, fileName).Scan(&docID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("create document: %w", err)
	}

	for _, chunk := range chunks {
		_, err := tx.Exec(ctx, `
		INSERT INTO document_chunks
			(chunk_index, content, embedding, document_id)
		VALUES ($1, $2, $3, $4)
	`,
			chunk.ChunkIndex,
			chunk.Content,
			pgvector.NewVector(chunk.Embedding),
			docID,
		)
		if err != nil {
			return uuid.Nil, fmt.Errorf("save chunk %d: %w", chunk.ChunkIndex, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit document transaction: %w", err)
	}

	return docID, nil
}

func (r *ChunkRepository) Search(ctx context.Context, embedding []float32, docId uuid.UUID) ([]SearchResult, error) {

	if len(embedding) != 3072 {
		return nil, fmt.Errorf(
			"expected 3072 embedding dimensions, got %d",
			len(embedding),
		)
	}

	rows, err := r.db.Query(ctx, `
        SELECT d.document_name, c.content
        FROM document_chunks c
        JOIN documents d ON d.id = c.document_id
        WHERE c.document_id = $2
        ORDER BY c.embedding <=> $1
        LIMIT 5
    `, pgvector.NewVector(embedding), docId)

	if err != nil {
		return nil, fmt.Errorf("search chunks: %w", err)
	}
	defer rows.Close()

	var results []SearchResult

	for rows.Next() {
		var result SearchResult

		if err := rows.Scan(
			&result.DocumentName,
			&result.Content,
		); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search results: %w", err)
	}

	return results, nil
}
