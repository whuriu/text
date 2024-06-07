package text_analysis

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestJsonPrepoc(t *testing.T) {
	// Create test directories if they do not exist
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("Error getting current directory: %v\n", err)
	}
	jsonDir := filepath.Join(root, "json_files")
	txtDir := filepath.Join(root, "txt_files")

	if err := os.MkdirAll(jsonDir, os.ModePerm); err != nil {
		t.Fatalf("Error creating json_files directory: %v\n", err)
	}
	if err := os.MkdirAll(txtDir, os.ModePerm); err != nil {
		t.Fatalf("Error creating txt_files directory: %v\n", err)
	}

	// Create a sample JSON file for testing
	filename := "preproc_test"
	jsonFilePath := filepath.Join(jsonDir, filename+".json")
	testData := map[string]interface{}{
		"messages": []map[string]interface{}{
			{"text": "Привіт, Світ!"},
			{"text": "Це тест."},
			{"text": "❤️🥺😏"},
		},
	}

	jsonFile, err := os.Create(jsonFilePath)
	if err != nil {
		t.Fatalf("Error creating JSON file: %v\n", err)
	}
	defer jsonFile.Close()

	if err := json.NewEncoder(jsonFile).Encode(testData); err != nil {
		t.Fatalf("Error writing to JSON file: %v\n", err)
	}

	// Expected string after cleaning
	expectedString := "привіт світ це тест"

	// Run the function
	if err := JsonPrepoc(filename); err != nil {
		t.Fatalf("JsonPrepoc failed: %v", err)
	}

	// Verify the output file content
	txtFilePath := filepath.Join(txtDir, filename+".txt")
	f, err := os.Open(txtFilePath)
	if err != nil {
		t.Fatalf("Opening file: %v\n", err)
	}
	defer f.Close()

	gotString, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("Reading file: %v\n", err)
	}

	if string(gotString) != expectedString {
		t.Fatalf("Expected %q, but got %q", expectedString, string(gotString))
	}
}
