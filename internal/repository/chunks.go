package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"

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
	DocID uuid.UUID
	ChunkIndex   int
	Content      string
	Embedding    []float32
}

type SearchResult struct {
	DocumentName string
	Content      string
}

func (r *ChunkRepository) CreateDocument(ctx context.Context,fileName string,) (uuid.UUID, error) {
    var docID uuid.UUID

    err := r.db.QueryRow(ctx, `
        INSERT INTO documents (document_name)
        VALUES ($1)
        RETURNING id
    `, fileName).Scan(&docID)

    if err != nil {
        return uuid.Nil, fmt.Errorf("create document: %w", err)
    }

    return docID, nil
}


func (r *ChunkRepository) Save(ctx context.Context, args SaveParameters) error {
	if len(args.Embedding) != 3072 {
		return fmt.Errorf(
			"expected 3072 embedding dimensions, got %d",
			len(args.Embedding),
		)
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO document_chunks
			(chunk_index, content, embedding,document_id)
		VALUES ($1, $2, $3, $4)
	`,
		args.ChunkIndex,
		args.Content,
		pgvector.NewVector(args.Embedding),
		args.DocID,
	)
	if err != nil {
		return fmt.Errorf("save chunk: %w", err)
	}

	return nil
}

func (r *ChunkRepository) Search(ctx context.Context, embedding []float32,docId uuid.UUID) ([]SearchResult, error) {

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
    `,pgvector.NewVector(embedding),docId)

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
