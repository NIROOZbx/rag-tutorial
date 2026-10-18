package document

import (
    "reflect"
    "testing"
)

func TestChunkText(t *testing.T) {
    tests := []struct {
        name      string
        text      string
        chunkSize int
        want      []string
    }{
        {
            name:      "normal text",
            text:      "I love learning Go",
            chunkSize: 2,
            want:      []string{"I love", "learning Go"},
        },
        {
            name:      "incomplete last chunk",
            text:      "hello you are",
            chunkSize: 2,
            want:      []string{"hello you", "are"},
        },
        {
            name:      "empty text",
            text:      "",
            chunkSize: 2,
            want:      nil,
        },
        {
            name:      "invalid chunk size",
            text:      "hello world",
            chunkSize: 0,
            want:      nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ChunkText(tt.text, tt.chunkSize)

            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}