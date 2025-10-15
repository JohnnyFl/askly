package services

import (
	"strings"
	"unicode"
)

const (
	// DefaultChunkSize is the target number of words per chunk
	DefaultChunkSize = 600

	// DefaultOverlapSize is the number of words to overlap between chunks
	DefaultOverlapSize = 100
)

// ChunkingService provides text chunking functionality
type ChunkingService struct {
	chunkSize   int
	overlapSize int
}

// NewChunkingService creates a new chunking service
func NewChunkingService() *ChunkingService {
	return &ChunkingService{
		chunkSize:   DefaultChunkSize,
		overlapSize: DefaultOverlapSize,
	}
}

// NewChunkingServiceWithSize creates a new chunking service with custom sizes
func NewChunkingServiceWithSize(chunkSize, overlapSize int) *ChunkingService {
	return &ChunkingService{
		chunkSize:   chunkSize,
		overlapSize: overlapSize,
	}
}

// ChunkText splits text into overlapping chunks of approximately chunkSize words
func (s *ChunkingService) ChunkText(text string) []string {
	// Normalize whitespace
	text = normalizeWhitespace(text)

	// Split into sentences
	sentences := splitIntoSentences(text)

	if len(sentences) == 0 {
		return []string{}
	}

	var chunks []string
	var currentChunk []string
	var currentWordCount int

	for _, sentence := range sentences {
		sentenceWords := countWords(sentence)

		// If adding this sentence would exceed chunk size and we have content
		if currentWordCount > 0 && currentWordCount+sentenceWords > s.chunkSize {
			// Save current chunk
			chunks = append(chunks, strings.Join(currentChunk, " "))

			// Create overlap from the end of current chunk
			overlapSentences := s.getOverlapSentences(currentChunk)
			currentChunk = overlapSentences
			currentWordCount = countWordsInSentences(currentChunk)
		}

		// Add sentence to current chunk
		currentChunk = append(currentChunk, sentence)
		currentWordCount += sentenceWords
	}

	// Add the last chunk if it has content
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}

	return chunks
}

// getOverlapSentences gets the last N sentences that fit within overlapSize words
func (s *ChunkingService) getOverlapSentences(sentences []string) []string {
	if s.overlapSize == 0 || len(sentences) == 0 {
		return []string{}
	}

	var overlap []string
	var wordCount int

	// Work backwards through sentences
	for i := len(sentences) - 1; i >= 0; i-- {
		sentence := sentences[i]
		sentenceWords := countWords(sentence)

		if wordCount+sentenceWords > s.overlapSize {
			break
		}

		overlap = append([]string{sentence}, overlap...)
		wordCount += sentenceWords
	}

	return overlap
}

// normalizeWhitespace normalizes whitespace in text
func normalizeWhitespace(text string) string {
	// Replace multiple spaces with single space
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

// splitIntoSentences splits text into sentences using basic rules
func splitIntoSentences(text string) []string {
	// Basic sentence splitting on . ! ? followed by space
	var sentences []string
	var current strings.Builder

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		current.WriteRune(runes[i])

		// Check for sentence ending
		if isSentenceEnd(runes, i) {
			sentence := strings.TrimSpace(current.String())
			if sentence != "" {
				sentences = append(sentences, sentence)
			}
			current.Reset()
		}
	}

	// Add any remaining text as final sentence
	if current.Len() > 0 {
		sentence := strings.TrimSpace(current.String())
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}

// isSentenceEnd checks if position i is a sentence ending
func isSentenceEnd(runes []rune, i int) bool {
	if i >= len(runes) {
		return false
	}

	r := runes[i]

	// Check if current char is sentence ending punctuation
	if r != '.' && r != '!' && r != '?' {
		return false
	}

	// If at end of text, it's a sentence end
	if i == len(runes)-1 {
		return true
	}

	// Check if followed by whitespace
	if i+1 < len(runes) && unicode.IsSpace(runes[i+1]) {
		// Check if next non-whitespace char is uppercase (new sentence)
		for j := i + 1; j < len(runes); j++ {
			if !unicode.IsSpace(runes[j]) {
				// If next char is uppercase or punctuation, it's likely a new sentence
				return unicode.IsUpper(runes[j]) || unicode.IsPunct(runes[j])
			}
		}
		return true
	}

	return false
}

// countWords counts the number of words in a string
func countWords(text string) int {
	if text == "" {
		return 0
	}
	return len(strings.Fields(text))
}

// countWordsInSentences counts total words in a slice of sentences
func countWordsInSentences(sentences []string) int {
	total := 0
	for _, sentence := range sentences {
		total += countWords(sentence)
	}
	return total
}

// GetChunkMetadata returns information about how text would be chunked
func (s *ChunkingService) GetChunkMetadata(text string) map[string]interface{} {
	chunks := s.ChunkText(text)

	var minWords, maxWords, totalWords int
	if len(chunks) > 0 {
		minWords = countWords(chunks[0])
		maxWords = minWords

		for _, chunk := range chunks {
			words := countWords(chunk)
			totalWords += words
			if words < minWords {
				minWords = words
			}
			if words > maxWords {
				maxWords = words
			}
		}
	}

	avgWords := 0
	if len(chunks) > 0 {
		avgWords = totalWords / len(chunks)
	}

	return map[string]interface{}{
		"chunk_count":         len(chunks),
		"total_words":         totalWords,
		"avg_words_per_chunk": avgWords,
		"min_words_per_chunk": minWords,
		"max_words_per_chunk": maxWords,
		"chunk_size_target":   s.chunkSize,
		"overlap_size":        s.overlapSize,
	}
}
