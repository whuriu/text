package text_analysis

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
)

type vocabItem struct {
	word string
	freq int
}

func TextAnalysis(filename string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	fp := filepath.Join(root, "text_analysis", "txt_files", filename+".txt")
	file, err := os.Open(fp)
	if err != nil {
		log.Fatalf("The file %s.txt was not found in the directory 'txt_files'.\n", filename)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("failed to close the file: %v\n", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	wordCount := make(map[string]int)

	for scanner.Scan() {
		word := scanner.Text()
		wordCount[word]++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	var vocab []vocabItem
	totalFreq := 0

	for word, freq := range wordCount {
		vocab = append(vocab, vocabItem{word, freq})
		totalFreq += freq
	}

	// Calculate normalized frequencies
	normFreq := make([]float64, len(vocab))
	for i, item := range vocab {
		normFreq[i] = float64(item.freq) / float64(totalFreq)
	}

	N := len(vocab)
	N1, N2 := 0, 0
	for _, item := range vocab {
		if item.freq == 1 {
			N1++
		} else if item.freq == 2 {
			N2++
		}
	}

	N_1N := float64(N1) / float64(N)
	N_2N := float64(N2) / float64(N)
	entropy := 0.0
	for _, nf := range normFreq {
		entropy -= nf * math.Log(nf)
	}

	fmt.Println("-------------------------")
	fmt.Printf("Довжина тексту: %d\n", N)
	fmt.Printf("Гапакс Легомена: %d\n", N1)
	fmt.Printf("Дис Легомена: %d\n", N2)
	fmt.Printf("Рівняння Гапакс Легомена: %.4f\n", N_1N)
	fmt.Printf("Рівняння Дис Легомена: %.4f\n", N_2N)
	fmt.Printf("Ентропія: %.4f\n", entropy)
	fmt.Println("-------------------------")
	return nil
}
