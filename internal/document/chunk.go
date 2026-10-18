package document

import (
	"strings"
)

func ChunkText(text string, size int) []string {

	words:=strings.Fields(text)

	if size==0|| len(words)==0{
		return nil
	}

	var chunks []string

	for i:=0;i<len(words);i+=size{
		end:=i+size

		if end>len(words){
			end=len(words)
		}

		chunk:=strings.Join(words[i:end]," ")
		chunks=append(chunks, chunk)
	}

	return chunks
}