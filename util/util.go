package util

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
)

// Function to convert a byte slice to base64 string
func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func imageToBase64(file string) string {
	// Read the entire file into a byte slice
	bytes, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	var base64Encoding string

	// Determine the content type of the image file
	mimeType := http.DetectContentType(bytes)

	// Prepend the appropriate URI scheme header depending on the MIME type
	switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding += "data:image/png;base64,"
	}

	// Append the base64 encoded output
	base64Encoding += toBase64(bytes)

	return base64Encoding
}
