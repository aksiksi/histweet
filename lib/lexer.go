package histweet

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
)

const (
	TIME_LAYOUT = "02-Jan-2006"
)

type Token struct {
	kind string
	val  string
	pos  int
	size int
}

// Internal lexer state
type Lexer struct {
	patterns  map[string]*regexp.Regexp
	input     string
	pos       int
	numTokens int
}

func NewLexer(tokens map[string]string, input string) *Lexer {
	lexer := &Lexer{
		patterns:  make(map[string]*regexp.Regexp),
		input:     strings.TrimSpace(input),
		pos:       0,
		numTokens: 0,
	}

	for k, v := range tokens {
		lexer.patterns[k] = regexp.MustCompile(v)
	}

	return lexer
}

// Fetches the next token from the input
// Returns an error if no valid token was found
func (lex *Lexer) PeekToken() (*Token, error) {
	// Iterate over each pattern and find the closest match
	// TODO: Can we improve this?
	matchPos := []int{math.MaxInt32, 0}
	matchType := ""

	token := &Token{}

	if lex.pos >= len(lex.input) {
		// Reached the end of the input
		token.kind = "EOF"
		token.pos = -1
		return token, nil
	}

	for k, v := range lex.patterns {
		// Check for a match
		location := v.FindStringIndex(lex.input[lex.pos:])
		if location == nil {
			continue
		}

		tmpMatchLen := location[1] - location[0]
		currMatchLen := matchPos[1] - matchPos[0]

		if location[0] < matchPos[0] {
			matchType = k
			matchPos = location
		} else if location[0] == matchPos[0] && tmpMatchLen > currMatchLen {
			// Always select the token with the longest match
			matchType = k
			matchPos = location
		}
	}

	if matchType == "" {
		msg := fmt.Sprintf("No match found at position %d", lex.pos)
		return nil, errors.New(msg)
	}

	start, end := matchPos[0], matchPos[1]
	matchLen := end - start

	// Update the token before returning it
	token.kind = matchType
	token.pos = lex.pos + start
	token.size = matchLen
	token.val = strings.TrimSpace(lex.input[token.pos : token.pos+matchLen])

	return token, nil
}

// Fetches next token in the input, then advances the lexer position
func (lex *Lexer) NextToken() (*Token, error) {
	token, err := lex.PeekToken()
	if err != nil {
		return token, err
	}

	// Advance the lexer position past the end of the match
	lex.pos = token.pos + token.size
	lex.numTokens++

	return token, nil
}

func (lex *Lexer) Reset() {
	lex.pos = 0
	lex.numTokens = 0
}
