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
	"strconv"
	"strings"
	"time"

	"github.com/aksiksi/histweet/lib"
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

type ParseNode struct {
	// Rule associated with this parse node
	rule *histweet.RuleTweet

	kind string

	op string

	// Logical operator to apply to the right of this parse node
	logical string

	// List of child nodes
	children []*ParseNode

	numChildren int
}

func (node *ParseNode) String() string {
	return fmt.Sprintf("Kind: %s, Logical: %s, Op: %s, NumChildren: %d, Rule: %+v",
		node.kind, node.logical, node.op, node.numChildren, node.rule)
}

// LL(1) parser for grammar defined above
type Parser struct {
	lexer *Lexer

	// Pointer to the current token
	currToken *Token

	// Tree of parse nodes
	treeRoot *ParseNode
}

func NewParser(input string) *Parser {
	// Map from token ID to regex pattern
	tokens := map[string]string{
		"IDENT":  "[a-zA-Z_]+",
		"NUMBER": "[0-9]+",
		"STRING": `"[^\"]*"`,
		"AGE":    `^\s*([0-9]+[ymd])?([0-9]+[ymd])?([0-9]+[ymd])`,
		"TIME":   `\d\d-\w\w\w-\d\d\d\d`,
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

	lexer := NewLexer(tokens, input)

	parser := &Parser{
		lexer: lexer,
		treeRoot: &ParseNode{
			children:    make([]*ParseNode, 0, 100),
			numChildren: 0,
			kind:        "root",
			rule:        nil,
		},
	}

	return parser
}

// Verifies that current token is of the specified `kind`,
// returns it, and reads in the next token
func (parser *Parser) match(kind string) (string, string, error) {
	currToken := parser.currToken
	val := currToken.val

	if currToken.kind != kind {
		msg := fmt.Sprintf("Failed match for kind = %s", kind)
		return "", "", errors.New(msg)
	}

	token, err := parser.lexer.NextToken()
	if err != nil {
		return "", "", err
	}

	parser.currToken = token

	return val, kind, nil
}

func (parser *Parser) expr(parent *ParseNode) (*ParseNode, error) {
	var op string
	var err error
	var node, logicalNode *ParseNode

	if parser.currToken.kind == "LPAREN" {
		_, _, err = parser.match("LPAREN")
		if err != nil {
			return nil, err
		}

		// For an expression, we know that this a non-terminal node,
		// so we construct a new "parent" for the expr here
		node = &ParseNode{
			children:    make([]*ParseNode, 0, 100),
			numChildren: 0,
			kind:        "expr",
			rule:        nil,
		}

		_, err = parser.expr(node)
		if err != nil {
			return nil, err
		}

		_, _, err = parser.match("RPAREN")
		if err != nil {
			return nil, err
		}
	} else {
		// Condition
		node, err = parser.cond()
		if err != nil {
			return nil, err
		}
	}

	// Improve this logic (TODO)
	op, err = parser.logical()
	if err == nil {
		// Logical found (TODO)
		// Build a logical node and make it the new "parent" for the node
		logicalNode = &ParseNode{
			children:    make([]*ParseNode, 0, 2),
			numChildren: 0,
			kind:        "logical",
			logical:     op,
			rule:        nil,
		}

		logicalNode.children = append(logicalNode.children, node)
		logicalNode.numChildren++

		_, err = parser.expr(logicalNode)
		if err != nil {
			return nil, err
		}

		node = logicalNode
	}

	// Insert this node into the current parent node
	parent.children = append(parent.children, node)
	parent.numChildren++

	return parent, nil
}

func (parser *Parser) cond() (*ParseNode, error) {
	ident, err := parser.ident()
	if err != nil {
		return nil, err
	}

	op, err1 := parser.op()
	if err1 != nil {
		return nil, err1
	}

	literal, kind, err2 := parser.literal()
	if err2 != nil {
		return nil, err2
	}

	// Build the rule
	rule := &histweet.RuleTweet{}

	switch ident {
	case "age":
		if kind != "AGE" {
			return nil, errors.New("Invalid literal for \"age\"")
		}

		time, err3 := ConvertAgeToTime(literal)
		if err3 != nil {
			return nil, err3
		}

		switch op {
		case "GT", "GTE":
			rule.Before = time
		case "LT", "LTE":
			rule.After = time
		default:
			return nil, errors.New("Invalid operator for \"age\"")
		}
	case "text":
		switch op {
		case "IN", "NOTIN":
			rule.Match = regexp.MustCompile(literal)
		default:
			return nil, errors.New("Invalid operator for \"text\"")
		}
	case "created":
		if kind != "TIME" {
			return nil, errors.New("Invalid literal for \"created\"")
		}

		time, err4 := time.Parse(TIME_LAYOUT, literal)
		if err4 != nil {
			return nil, errors.New(fmt.Sprintf("Invalid time provided: %s", literal))
		}

		switch op {
		case "GT", "GTE":
			rule.Before = time
		case "LT", "LTE":
			rule.After = time
		default:
			return nil, errors.New(fmt.Sprintf("Invalid time comparison operation: %s", op))
		}
	case "likes":
		switch op {
		case "LT", "LTE", "GT", "GTE", "EQ", "NEQ":
			num, err := strconv.Atoi(literal)
			if err != nil {
				return nil, err
			}

			rule.MaxLikes = num
		default:
			return nil, errors.New("Invalid operator for \"likes\"")
		}
	default:
		return nil, errors.New(fmt.Sprintf("Invalid identifier \"%s\"", ident))
	}

	node := &ParseNode{
		kind:     "cond",
		rule:     rule,
		op:       op,
		children: nil,
	}

	return node, nil
}

func (parser *Parser) ident() (string, error) {
	val, _, err := parser.match("IDENT")
	if err != nil {
		return "", err
	}

	return val, nil
}

func (parser *Parser) logical() (string, error) {
	var kind string
	var err error

	switch parser.currToken.kind {
	case "AND", "OR":
		_, kind, err = parser.match(parser.currToken.kind)
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("Invalid logical operator")
	}

	return kind, nil
}

func (parser *Parser) op() (string, error) {
	var kind string
	var err error

	switch parser.currToken.kind {
	case "GTE", "GT", "LTE", "LT", "EQ", "NEQ", "IN", "NOTIN":
		_, kind, err = parser.match(parser.currToken.kind)
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("Invalid comparison operator")
	}

	return kind, nil
}

func (parser *Parser) literal() (string, string, error) {
	var val string
	var kind string
	var err error

	switch parser.currToken.kind {
	case "STRING", "NUMBER", "AGE", "TIME":
		val, kind, err = parser.match(parser.currToken.kind)
		if err != nil {
			return "", "", err
		}
	default:
		return "", "", errors.New("Invalid literal")
	}

	return val, kind, nil
}

func PrintParseTree(currNode *ParseNode, depth int) {
	if currNode.numChildren == 0 {
		return
	}

	for _, node := range currNode.children {
		fmt.Printf("depth = %d, %s\n", depth, node)
		PrintParseTree(node, depth+1)
	}
}

// Entry point for parser
func (parser *Parser) Parse() error {
	token, err := parser.lexer.NextToken()
	if err != nil {
		return err
	}

	parser.currToken = token

	_, err1 := parser.expr(parser.treeRoot)

	// PrintParseTree(parser.treeRoot, 0)

	return err1
}
