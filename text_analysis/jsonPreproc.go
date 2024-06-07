package text_analysis

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

func JsonPrepoc(filename string) error {
	root, err := filepath.Abs(".")
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return err
	}

	jsonFilePath := filepath.Join(root, "text_analysis", "json_files", fmt.Sprintf("%s.json", filename))
	fJson, err := os.Open(jsonFilePath)
	if err != nil {
		return err
	}
	defer fJson.Close()
	jsonFile, err := io.ReadAll(fJson)
	if err != nil {
		fmt.Printf("Error reading JSON file: %v\n", err)
		return err
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(jsonFile, &obj); err != nil {
		fmt.Printf("Error unmarshalling JSON file: %v\n", err)
		return err
	}
	var texts []string
	if messages, ok := obj["messages"].([]interface{}); ok {
		for _, message := range messages {
			if msgMap, ok := message.(map[string]interface{}); ok {
				if text, ok := msgMap["text"].(string); ok {
					texts = append(texts, text)
				}
			}
		}
	}
	combinedText := strings.Join(texts, " ")

	combinedText = cleanString(combinedText)
	txtFilePath := filepath.Join(root, "text_analysis", "txt_files", fmt.Sprintf("%s.txt", filename))
	fTxt, errTXT := os.OpenFile(txtFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if errTXT != nil {
		return errTXT
	}
	if _, err := fTxt.Write([]byte(combinedText)); err != nil {
		fmt.Printf("Error writing text file: %v\n", err)
		return err
	}

	fmt.Println("Text processing completed successfully.")
	return nil
}

// Function to check if a rune is an emoji
func isEmoji(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Miscellaneous Symbols and Pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map Symbols
		(r >= 0x2600 && r <= 0x26FF) || // Miscellaneous Symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0xFE00 && r <= 0xFE0F) || // Variation Selectors
		(r >= 0x1F900 && r <= 0x1F9FF) // Supplemental Symbols and Pictographs
}

func cleanString(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, ",", "")
	input = strings.ReplaceAll(input, ".", "")
	input = strings.ReplaceAll(input, `"`, "")
	input = strings.ReplaceAll(input, ":", " ")
	input = strings.ReplaceAll(input, "–", " ")
	input = strings.ReplaceAll(input, "(", "")
	input = strings.ReplaceAll(input, ")", "")
	input = strings.ReplaceAll(input, ";", "")
	input = strings.ReplaceAll(input, "!", "")
	input = strings.ReplaceAll(input, "?", "")
	input = strings.ReplaceAll(input, "*", "")

	result := make([]rune, 0, len(input))
	i := 0
	for _, r := range input {
		i++
		if isEmoji(r) {
			result = append(result, ' ')
		} else if unicode.Is(unicode.Cyrillic, r) || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			result = append(result, r)
		}
	}
	input = string(result)
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")
	input = strings.ToLower(input)
	input = strings.TrimSpace(input)
	return input
}
