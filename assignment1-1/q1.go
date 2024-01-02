package cos418_hw1_1

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

// Find the top K most common words in a text document.
//
//	path: location of the document
//	numWords: number of words to return (i.e. k)
//	charThreshold: character threshold for whether a token qualifies as a word,
//		e.g. charThreshold = 5 means "apple" is a word but "pear" is not.
//
// Matching is case insensitive, e.g. "Orange" and "orange" is considered the same word.
// A word comprises alphanumeric characters only. All punctuation and other characters
// are removed, e.g. "don't" becomes "dont".
// You should use `checkError` to handle potential errors.
func topWords(path string, numWords int, charThreshold int) []WordCount {
	m := make(map[string]int)
	//reg := regexp.MustCompile(`^(.*/)?(?:$|(.+?)(?:(\.[^.]*$)|$))`)
	//s := reg.FindStringSubmatch(path)

	b, err := os.ReadFile(path)
	s := string(b) // converting from bytes to string

	w := []WordCount{}
	re, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	wordlist := strings.Fields(s)

	for i, v := range wordlist {
		wordlist[i] = strings.ToLower(v)
		wordlist[i] = re.ReplaceAllString(wordlist[i], "")
	}
	//fmt.Println("Printing", wordlist)
	for _, v := range wordlist {
		if len(v) >= charThreshold {
			m[v]++
		}
	}
	for i, v := range m {
		w = append(w, WordCount{i, v})
	}
	sortWordCounts(w)
	// HINT: You may find the `strings.Fields` and `strings.ToLower` functions helpful
	// HINT: To keep only alphanumeric characters, use the regex "[^0-9a-zA-Z]+"
	return w[:numWords]
}

// A struct that represents how many times a word is observed in a document
type WordCount struct {
	Word  string
	Count int
}

func (wc WordCount) String() string {
	return fmt.Sprintf("%v: %v", wc.Word, wc.Count)
}

// Helper function to sort a list of word counts in place.
// This sorts by the count in decreasing order, breaking ties using the word.
// DO NOT MODIFY THIS FUNCTION!
func sortWordCounts(wordCounts []WordCount) {
	sort.Slice(wordCounts, func(i, j int) bool {
		wc1 := wordCounts[i]
		wc2 := wordCounts[j]
		if wc1.Count == wc2.Count {
			return wc1.Word < wc2.Word
		}
		return wc1.Count > wc2.Count
	})
}
