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
 *   Expr     	  <-  ( Cond ) | Cond
 *	 Cond	      <-  Ident Op Literal (Logical Expr)?
 *   Logical      <-  Or | And
 *   Op           <-  Gt | Gte | Lt | Lte | Eq | Neq | In | NotIn
 *   Literal      <-  Number | String
 *
 *   Ident	      :=  [A-Za-z0-9_]+
 *   Number	      :=  [0-9]+
 *   String	      :=  " [^"]* "
 *	 Lparen       :=  (
 *	 Rparen       :=  )
 *   Or		      :=  ||
 *   And	      :=  &&
 * 	 Gt           :=  >
 * 	 Gte          :=  >=
 * 	 Lt           :=  <
 * 	 Lte          :=  <=
 * 	 Eq           :=  ==
 * 	 Neq          :=  !=
 * 	 In           :=  ~
 * 	 NotIn        :=  !~
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

func NewLexer(tokens map[string]string, input string, skipWhitespace bool) *Lexer {
	lexer := &Lexer{
		patterns:       make(map[string]*regexp.Regexp),
		input:          strings.TrimSpace(input),
		pos:            0,
		numTokens:      0,
		skipWhitespace: skipWhitespace,
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
		return nil, errors.New(msg)
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

// LL(1) parser for grammar defined above
type Parser struct {
	lexer *Lexer

	// Pointer to the current token
	currToken *Token
}

func NewParser(input string) *Parser {
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

	lexer := NewLexer(tokens, input, true)

	parser := &Parser{
		lexer: lexer,
	}

	return parser
}

// Verifies that current token is of the specified `kind`,
// returns it, and reads in the next token
func (parser *Parser) match(kind string) (string, error) {
	currToken := parser.currToken
	val := currToken.val

	fmt.Println(currToken)

	if currToken.kind != kind {
		msg := fmt.Sprintf("Failed match for kind = %s", kind)
		return "", errors.New(msg)
	}

	token, err := parser.lexer.NextToken()
	if err != nil {
		return "", err
	}

	parser.currToken = token

	return val, nil
}

// *   Expr     	<-  ( Expr ) (Logical Expr)? | Cond (Logical Expr)?
// *   Cond	        <-  Ident Op Literal
// *   Logical      <-  Or | And
// *   Op           <-  Gt | Gte | Lt | Lte | Eq | Neq | In | NotIn
// *   Literal      <-  Number | String
func (parser *Parser) expr() (bool, error) {
	if parser.currToken.kind == "LPAREN" {
		_, err := parser.match("LPAREN")
		if err != nil {
			return false, err
		}

		_, err1 := parser.expr()
		if err1 != nil {
			return false, err1
		}

		_, err2 := parser.match("RPAREN")
		if err2 != nil {
			return false, err2
		}
	} else {
		// Condition
		_, err := parser.cond()
		if err != nil {
			return false, err
		}
	}

	// Improve this logic (TODO)
	_, err3 := parser.logical()
	if err3 != nil {
		// No logical found (TODO)
		return true, nil
	}

	return parser.expr()
}

func (parser *Parser) cond() (bool, error) {
	_, err := parser.ident()
	if err != nil {
		return false, err
	}

	_, err1 := parser.op()
	if err1 != nil {
		return false, err1
	}

	_, err2 := parser.literal()
	if err2 != nil {
		return false, err2
	}

	return true, nil
}

func (parser *Parser) ident() (bool, error) {
	_, err := parser.match("IDENT")
	if err != nil {
		return false, err
	}

	return true, nil
}

func (parser *Parser) logical() (bool, error) {
	switch parser.currToken.kind {
	case "AND", "OR":
		_, err := parser.match(parser.currToken.kind)
		if err != nil {
			return false, err
		}
	default:
		return false, errors.New("Invalid logical operator")
	}

	return true, nil
}

func (parser *Parser) op() (bool, error) {
	switch parser.currToken.kind {
	case "GTE", "GT", "LTE", "LT", "EQ", "NEQ", "IN", "NOTIN":
		_, err := parser.match(parser.currToken.kind)
		if err != nil {
			return false, err
		}
	default:
		return false, errors.New("Invalid comparison operator")
	}

	return true, nil
}

func (parser *Parser) literal() (bool, error) {
	switch parser.currToken.kind {
	case "STRING", "NUMBER":
		_, err := parser.match(parser.currToken.kind)
		if err != nil {
			return false, err
		}
	default:
		return false, errors.New("Invalid literal")
	}

	return true, nil
}

// Entry point for parser
func (parser *Parser) Parse() error {
	token, err := parser.lexer.NextToken()
	if err != nil {
		return err
	}

	parser.currToken = token

	_, err1 := parser.expr()

	return err1
}
