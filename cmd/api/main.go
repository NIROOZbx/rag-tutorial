package main

import (
	"context"
	"log"
	"net/http"

	"github.com/NIROOZbx/rag-tutorial.git/internal/embedding"
	"github.com/NIROOZbx/rag-tutorial.git/internal/server"
)

func main() {
	mux := server.New()

	ctx := context.Background()

	_, err := embedding.NewGeminiClient(ctx)

	if err != nil {
		log.Println("Error occured", err)
	}



	log.Fatal(http.ListenAndServe(":8080", mux))

}
