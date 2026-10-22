# PDF RAG API

A Go API for uploading text-based PDFs, indexing their contents, and asking questions against a selected document. It extracts PDF text, splits it into chunks, creates Gemini embeddings, stores the chunks in Supabase Postgres with pgvector, then retrieves relevant chunks and asks Groq to generate an answer.

## How it works

1. `POST /upload` accepts a PDF as multipart form data.
2. The API extracts and cleans the text, divides it into chunks of 500 words, and creates a Gemini embedding for each chunk.
3. The document and chunks are saved together in a Postgres transaction.
4. `POST /question` embeds a question, finds the five closest chunks for the supplied document, and sends that context to Groq.

## Requirements

- Go version declared in [`go.mod`](go.mod) or newer.
- A Gemini API key with embedding model access.
- A Supabase Postgres database with the `vector` extension enabled.
- A Groq API key.

## Configuration

Set these environment variables in your shell, a local `.env` file, or your Render service environment. `.env` is loaded locally if present; the file is not required in production.

| Variable | Required | Purpose |
| --- | --- | --- |
| `GEMINI_API_KEY` | Yes | Used by the Google GenAI SDK to create embeddings with `gemini-embedding-2`. |
| `GROQ_API_KEY` | Yes | Used to generate question answers. |
| `SUPABASE_DB_URL` | Yes | Postgres connection URL for Supabase. |
| `PORT` | Render | Port to bind on Render. Defaults to `8080` for local development. |

Example `.env` (keep it private; never commit real credentials):

```env
GEMINI_API_KEY=your_gemini_api_key
GROQ_API_KEY=your_groq_api_key
SUPABASE_DB_URL=postgresql://user:password@host:5432/postgres
PORT=8080
```

## Database setup

Run this once in the Supabase SQL Editor. The vector dimension must remain `3072` to match the embeddings and repository checks in this project.

```sql
create extension if not exists vector;

create table if not exists documents (
  id uuid primary key default gen_random_uuid(),
  document_name text not null,
  created_at timestamptz not null default now()
);

create table if not exists document_chunks (
  id bigint generated always as identity primary key,
  document_id uuid not null references documents(id) on delete cascade,
  chunk_index integer not null,
  content text not null,
  embedding vector(3072) not null,
  unique (document_id, chunk_index)
);

create index if not exists document_chunks_document_id_idx
  on document_chunks (document_id);
```

The app connects directly to Postgres using `SUPABASE_DB_URL`; use the connection string appropriate for your deployment/network and keep its password secret.

## Run locally

From the repository root:

```sh
go run ./cmd/api
```

The API listens on `PORT` when set, or `http://localhost:8080` otherwise. Check it with:

```sh
curl http://localhost:8080/health
```

## API

### `GET /health`

Returns a small health response when the server is reachable.

```json
{"data":"Health check ok"}
```

### `POST /upload`

Send multipart form data with a file field named **`document`**. The total request body is limited to 10 MiB. Upload a text-based PDF; image-only scans may have no extractable text unless OCR has been run beforehand.

```sh
curl -X POST http://localhost:8080/upload \
  -F "document=@./example.pdf"
```

Successful response:

```json
{
  "message": "PDF processed successfully",
  "document_id": "00000000-0000-0000-0000-000000000000",
  "document_name": "example.pdf",
  "chunks": 12
}
```

Save the returned `document_id` to ask questions about that PDF.

### `POST /question`

Send JSON with a question and the document ID returned by `/upload`:

```sh
curl -X POST http://localhost:8080/question \
  -H "Content-Type: application/json" \
  -d '{"question":"What is this document about?","doc_id":"00000000-0000-0000-0000-000000000000"}'
```

Successful response:

```json
{"answer":"The document is about ..."}
```

The API returns `400` for an invalid request, `413` when the upload exceeds the size limit, and `500` when PDF processing, embedding, database storage, or answer generation fails. Detailed upload failure causes are written to the server logs.

## Deploy on Render

Create a Render **Web Service** for this repository and configure:

- **Build command:** `go build -o app ./cmd/api`
- **Start command:** `./app`
- **Environment variables:** `GEMINI_API_KEY`, `GROQ_API_KEY`, and `SUPABASE_DB_URL`

Render supplies `PORT`; the app binds to it automatically. Check the Render service logs for startup errors and upload diagnostics.

## CORS

The server currently allows browser requests only from `https://doc-reader.nirooz.xyz`. If the frontend origin changes, update the origin allowlist in `internal/server/server.go` and redeploy. CORS controls whether browsers can read API responses; it does not fix server-side upload or ingestion errors.

## Troubleshooting uploads

- **`413` / file-size message:** the request exceeds the 10 MiB limit.
- **No extractable text:** the PDF may be scanned images; run OCR before uploading.
- **Gemini `context canceled`:** the upload was interrupted while embeddings were being generated. Keep the upload page open and check client-side timeouts or abort logic.
- **Postgres invalid UTF-8 / NUL error:** the extractor cleans NUL bytes and invalid UTF-8 before processing; redeploy the commit containing that fix.
- **Other `500` responses:** inspect Render logs for `PDF processing failed`; the wrapped error identifies PDF extraction, Gemini embedding, or database storage as the failing stage.

## Project layout

```text
cmd/api/                 Application entry point
internal/document/       PDF text extraction and chunking
internal/embedding/      Gemini embedding client
internal/groq/           Groq answer generation client
internal/database/       Supabase Postgres connection
internal/repository/     Document chunk persistence and vector search
internal/service/        Ingestion and question answering workflows
internal/handler/        HTTP handlers
internal/server/         Routes and CORS middleware
internal/worker/         PDF extraction worker pool
```
