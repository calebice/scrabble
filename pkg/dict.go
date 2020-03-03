package scrabble

import (
	"bufio"
	"os"
)

// Dictionary represents the presence of a word in the scrabble dictionary
type Dictionary struct {
	Words map[string]bool
}

// LoadDictionary Opens the path to a line separated dictionary and builds a working
// game dictionary
func LoadDictionary(path string) (Dictionary, error) {
	var Dict Dictionary
	Dict.Words = make(map[string]bool)
	file, err := os.Open(path)
	if err != nil {
		return Dict, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		Dict.Words[scanner.Text()] = true
	}

	return Dict, nil
}
