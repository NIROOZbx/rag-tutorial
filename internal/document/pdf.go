package document

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ledongthuc/pdf"
)

func ExtractPDFText(file io.Reader)(string,error){

	data,err:=io.ReadAll(file)

	if err != nil {
		return "", fmt.Errorf("read uploaded PDF: %w", err)
	}

	reader, err := pdf.NewReader(
		bytes.NewReader(data),
		int64(len(data)),
	)
	if err != nil {
		return "", fmt.Errorf("open PDF: %w", err)
	}

	textReader,err:=reader.GetPlainText()

	if err != nil {
		return "", fmt.Errorf("extract PDF text: %w", err)
	}

	text, err := io.ReadAll(textReader)
	if err != nil {
		return "", fmt.Errorf("read extracted text: %w", err)
	}


	return string(text), nil
	
}
