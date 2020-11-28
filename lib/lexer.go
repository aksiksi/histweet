package histweet

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

type tokenKind int

const (
	// Identifiers and literals
	tokenIdent tokenKind = iota
	tokenNumber
	tokenString
	tokenAge
	tokenTime

	// Grouping
	tokenLparen
	tokenRparen

	// Logical operators
	tokenOr
	tokenAnd

	// Comparison operators
	tokenGte
	tokenGt
	tokenLte
	tokenLt
	tokenEq
	tokenNeq
	tokenIn
	tokenNotIn

	tokenEOF
)

func (t tokenKind) ToString() string {
	switch t {
	case tokenIdent:
		return "identifier"
	case tokenNumber:
		return "number"
	case tokenString:
		return "string"
	case tokenAge:
		return "age"
	case tokenTime:
		return "time"
	case tokenLparen:
		return "left paren"
	case tokenRparen:
		return "right paren"
	case tokenOr:
		return "or"
	case tokenAnd:
		return "and"
	case tokenGte:
		return "greater or equal"
	case tokenGt:
		return "greater"
	case tokenLte:
		return "less or equal"
	case tokenLt:
		return "less"
	case tokenEq:
		return "equal"
	case tokenNeq:
		return "not equal"
	case tokenIn:
		return "in"
	case tokenNotIn:
		return "not in"
	case tokenEOF:
		return "eof"
	default:
		return "unknown"
	}
}

type token struct {
	kind tokenKind
	val  string
	pos  int
	size int
}

type lexer struct {
	patterns  map[tokenKind]*regexp.Regexp
	input     string
	pos       int
	numTokens int
}

func newLexer(tokens map[tokenKind]string, input string) *lexer {
	lexer := &lexer{
		patterns:  make(map[tokenKind]*regexp.Regexp),
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
func (lex *lexer) peekToken() (*token, error) {
	// By default, token is EOF with a position one past the end of the input
	token := &token{kind: tokenEOF, pos: len(lex.input)}

	if lex.pos >= len(lex.input) {
		// Reached the end of the input
		return token, nil
	}

	matchPos := []int{math.MaxInt32, 0}
	matchType := tokenEOF

	// Consume any whitespace characters in the input
	for {
		if lex.input[lex.pos] != ' ' {
			break
		}

		lex.pos++
	}

	// Iterate over each pattern and find the closest match
	// TODO: Can we improve this?
	for k, v := range lex.patterns {
		// Check for a match from the _start_ of the current
		// position in the stream. We enforce this by having
		// all token regexes start with "^".
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
			// Always select the token with the _longest_ match
			matchType = k
			matchPos = location
		}
	}

	if matchType == tokenEOF {
		return nil, fmt.Errorf("No match found at position %d", lex.pos)
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

// Fetch/peek the next token in the input, then advance the lexer position
func (lex *lexer) nextToken() (*token, error) {
	token, err := lex.peekToken()
	if err != nil {
		return token, err
	}

	// Advance the lexer position past the end of the match
	lex.pos = token.pos + token.size
	lex.numTokens++

	return token, nil
}

// Reset the lexer to the start of the input
func (lex *lexer) Reset() {
	lex.pos = 0
	lex.numTokens = 0
}
