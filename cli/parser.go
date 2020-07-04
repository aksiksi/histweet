/*
 * A simple parser for tweet deletion rules.
 *
 * Examples:
 *
 * - age > 3d
 * - age > 10m3d || likes == 0
 * - (likes > 10 && retweets > 3) || (text ~ "hello, world!")
 * - retweets >= 3 && time <= "10 May 2020"
 *
 * Grammar:
 *
 *   Rule     <-  Grouping | Expr | Cond
 *   Grouping <-  Lparen Expr Rparen
 *   Expr     <-  Cond Logical Cond
 *   Cond     <-  Ident Op Literal
 *   Logical  <-  Or | And
 *   Op       <-  Gt | Gte | Lt | Lte | Eq | Neq | In | NotIn
 *   Literal  <-  Number | String

 *   Ident	  :=  [A-Za-z0-9_]+
 *   Number	  :=  [0-9]+
 *   String	  :=  " [^"]* "
 *	 Lparen   :=  (
 *	 Rparen   :=  )
 *   Or		  :=  ||
 *   And	  :=  &&
 * 	 Gt       :=  >
 * 	 Gte      :=  >=
 * 	 Lt       :=  <
 * 	 Lte      :=  <=
 * 	 Eq       :=  ==
 * 	 Neq      :=  !=
 * 	 In       :=  ~
 * 	 NotIn    :=  !~
 */

package main

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
)

type Token struct {
	kind string
	val  string
	pos  int
	size int
}

// Internal lexer state
type Lexer struct {
	patterns       map[string]*regexp.Regexp
	input          string
	pos            int
	numTokens      int
	skipWhitespace bool
}

func NewLexer(input string, skipWhitespace bool) *Lexer {
	lexer := &Lexer{
		patterns:       make(map[string]*regexp.Regexp),
		input:          strings.TrimSpace(input),
		pos:            0,
		numTokens:      0,
		skipWhitespace: skipWhitespace,
	}

	// Map from token ID to regex pattern
	tokens := map[string]string{
		"IDENT":  "[a-zA-Z_]+",
		"NUMBER": "[0-9]+",
		"STRING": `"[^\"]*"`,
		"LPAREN": `\(`,
		"RPAREN": `\)`,
		"OR":     `\|\|`,
		"AND":    "&&",
		"GTE":    ">=",
		"GT":     ">",
		"LTE":    "<=",
		"LT":     "<",
		"EQ":     "==",
		"NEQ":    "!=",
		"IN":     "~",
		"NOTIN":  "!~",
	}

	for k, v := range tokens {
		lexer.patterns[k] = regexp.MustCompile(v)
	}

	return lexer
}

// Fetches the next token from the input
// Returns an error if no valid token was found
func (lex *Lexer) PeekToken() (Token, error) {
	// Iterate over each pattern and find the closest match
	// TODO: Can we improve this?
	matchPos := []int{math.MaxInt32, 0}
	matchType := ""

	var token Token

	if lex.pos >= len(lex.input) {
		// Reached the end of the input
		token.kind = "EOF"
		token.pos = -1
		return token, nil
	}

	for k, v := range lex.patterns {
		// TODO: Skip whitespace

		// Check for a match
		location := v.FindStringIndex(lex.input[lex.pos:])
		if location == nil {
			continue
		}

		if location[0] < matchPos[0] {
			matchType = k
			matchPos = location
		}
	}

	if matchType == "" {
		msg := fmt.Sprintf("No match found at position %d", lex.pos)
		return token, errors.New(msg)
	}

	start, end := matchPos[0], matchPos[1]
	matchLen := end - start

	// Update the token before returning it
	token.kind = matchType
	token.pos = lex.pos + start
	token.size = matchLen
	token.val = lex.input[token.pos : token.pos+matchLen]

	return token, nil
}

// Fetches next token in the input, then advances the lexer position
func (lex *Lexer) NextToken() (Token, error) {
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
