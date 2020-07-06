package histweet

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Tokens for terminals of the Twitter rule parser grammar
var TOKENS = map[string]string{
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

// Represents a single node in the parse tree.
//
// Each node has a kind, which is one of: "expr", "logical", or "cond".
// Logical nodes indicate that the node's two children are connected by a
// logical operation (&& or ||). Expr nodes indicate one or more expressions or
// conditions, tied by logical operators.
//
// If the node is a condition (cond) node, the rule field will contains the logic
// required to evaluate a match for a given tweet.
//
// Parsers can only be used once; to re-use a parser, make sure to call the
// Reset() method.
type ParseNode struct {
	kind        string
	op          string
	rule        *RuleTweet
	children    []*ParseNode
	numChildren int
}

func (node *ParseNode) String() string {
	return fmt.Sprintf("Kind: %s, Op: %s, NumChildren: %d, Rule: %+v",
		node.kind, node.op, node.numChildren, node.rule)
}

type ParsedRule struct {
	root     *ParseNode
	input    string
	numNodes int
}

func NewParsedRule(input string) *ParsedRule {
	root := &ParseNode{
		children:    make([]*ParseNode, 0, 100),
		numChildren: 0,
		kind:        "root",
		rule:        nil,
	}

	return &ParsedRule{
		root:     root,
		input:    input,
		numNodes: 0,
	}
}

func evalInternal(tweet *Tweet, node *ParseNode) bool {
	if node.kind == "cond" {
		return tweet.IsMatch(node.rule)
	} else if node.kind == "logical" {
		left := evalInternal(tweet, node.children[0])
		right := evalInternal(tweet, node.children[1])

		switch node.op {
		case "AND":
			return left && right
		case "OR":
			return left || right
		default:
			panic(fmt.Sprintf("Unexpected logical op: %s\n", node.op))
		}
	} else {
		// TODO: Does this make sense?
		return evalInternal(tweet, node.children[0])
	}
}

// Walk the parse tree and evaluate each condition against the given Tweet.
// Return true if the Tweet matches all of the rules.
func (rule *ParsedRule) IsMatch(tweet *Tweet) bool {
	root := rule.root
	return evalInternal(tweet, root)
}

// A simple parser for tweet deletion rules.
//
// Examples:
//
// - age > 3d
// - age > 10m3d || likes == 0
// - (likes > 10 && retweets > 3) || (text ~ "hello, world!")
// - retweets >= 3 && time <= "10 May 2020"
//
// Grammar:
//
// Expr    <-  ( Expr ) [Logical Expr]? | Cond [Logical Expr]?
// Cond	   <-  Ident Op Literal
// Logical <-  Or | And
// Op      <-  Gt | Gte | Lt | Lte | Eq | Neq | In | NotIn
// Literal <-  Number | String | Age | Time
//
// Ident   :=  [A-Za-z0-9_]+
// Number  :=  [0-9]+
// String  :=  " [^"]* "
// Age     :=  ^\s*([0-9]+[ymd])?([0-9]+[ymd])?([0-9]+[ymd])
// Time    :=  \d\d-\w\w\w-\d\d\d\d
// Lparen  :=  (
// Rparen  :=  )
// Or	   :=  ||
// And	   :=  &&
// Gt      :=  >
// Gte     :=  >=
// Lt      :=  <
// Lte     :=  <=
// Eq      :=  ==
// Neq     :=  !=
// In      :=  ~
// NotIn   :=  !~
type Parser struct {
	lexer *Lexer

	// Pointer to the current token
	currToken *Token

	// Tree of parse nodes
	rule *ParsedRule
}

type ParserError struct {
	msg   string
	token *Token
}

func (err *ParserError) Error() string {
	return fmt.Sprintf("Parser Error: %s (col %d)", err.msg, err.token.pos)
}

func NewParserError(msg string, token *Token) *ParserError {
	return &ParserError{
		msg:   msg,
		token: token,
	}
}

func NewParser(input string) *Parser {
	lexer := NewLexer(TOKENS, input)

	parser := &Parser{
		lexer: lexer,
		rule:  NewParsedRule(input),
	}

	return parser
}

// Verifies that current token is of the specified `kind`,
// returns it, and reads in the next token
func (parser *Parser) match(kind string) (*Token, error) {
	currToken := parser.currToken

	if currToken.kind != kind {
		msg := fmt.Sprintf("Failed match for kind = %s", kind)
		return nil, errors.New(msg)
	}

	token, err := parser.lexer.NextToken()
	if err != nil {
		return nil, err
	}

	parser.currToken = token

	return currToken, nil
}

func (parser *Parser) expr(parent *ParseNode) (*ParseNode, error) {
	var token *Token
	var err error
	var node, logicalNode *ParseNode

	if parser.currToken.kind == "LPAREN" {
		_, err = parser.match("LPAREN")
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

		_, err = parser.match("RPAREN")
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
	token, err = parser.logical()
	if token != nil {
		// Logical found (TODO)
		// Build a logical node and make it the new "parent" for the node
		logicalNode = &ParseNode{
			children:    make([]*ParseNode, 0, 2),
			numChildren: 0,
			kind:        "logical",
			op:          token.kind,
			rule:        nil,
		}

		logicalNode.children = append(logicalNode.children, node)
		logicalNode.numChildren++

		parser.rule.numNodes++

		_, err = parser.expr(logicalNode)
		if err != nil {
			return nil, err
		}

		node = logicalNode
	}

	// Insert this node into the current parent node
	parent.children = append(parent.children, node)
	parent.numChildren++

	parser.rule.numNodes++

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

	literal, err2 := parser.literal()
	if err2 != nil {
		return nil, err2
	}

	// Build the rule
	rule := &RuleTweet{}

	switch ident.val {
	case "age":
		if literal.kind != "AGE" {
			msg := fmt.Sprintf("Invalid literal for \"age\": %s", literal.val)
			return nil, NewParserError(msg, literal)
		}

		time, err3 := ConvertAgeToTime(literal.val)
		if err3 != nil {
			msg := fmt.Sprintf("Invalid format for \"age\": %s", literal.val)
			return nil, NewParserError(msg, literal)
		}

		switch op.kind {
		case "GT", "GTE":
			rule.Before = time
		case "LT", "LTE":
			rule.After = time
		default:
			msg := fmt.Sprintf("Invalid operator for \"age\": %s", op.val)
			return nil, NewParserError(msg, op)
		}
	case "text":
		switch op.kind {
		case "IN", "NOTIN":
			rule.Match = regexp.MustCompile(literal.val)
		default:
			msg := fmt.Sprintf("Invalid operator for \"text\": %s", op.val)
			return nil, NewParserError(msg, op)
		}
	case "created":
		if literal.kind != "TIME" {
			msg := fmt.Sprintf("Invalid literal for \"created\": %s", literal.val)
			return nil, NewParserError(msg, literal)
		}

		time, err4 := time.Parse(TIME_LAYOUT, literal.val)
		if err4 != nil {
			msg := fmt.Sprintf("Invalid format for time for \"created\": %s", literal.val)
			return nil, NewParserError(msg, literal)
		}

		switch op.kind {
		case "GT", "GTE":
			rule.Before = time
		case "LT", "LTE":
			rule.After = time
		default:
			msg := fmt.Sprintf("Invalid operator for \"created\": %s", op.val)
			return nil, NewParserError(msg, op)
		}
	case "likes":
		switch op.kind {
		case "LT", "LTE", "GT", "GTE", "EQ", "NEQ":
			num, err := strconv.Atoi(literal.val)
			if err != nil {
				msg := fmt.Sprintf("Invalid number for \"likes\": %s", literal.val)
				return nil, NewParserError(msg, literal)
			}

			// TODO
			rule.MaxLikes = num
		default:
			msg := fmt.Sprintf("Invalid operator for \"likes\": %s", op.val)
			return nil, NewParserError(msg, op)
		}
	case "reweets":
		switch op.kind {
		case "LT", "LTE", "GT", "GTE", "EQ", "NEQ":
			num, err := strconv.Atoi(literal.val)
			if err != nil {
				msg := fmt.Sprintf("Invalid number for \"reweets\": %s", literal.val)
				return nil, NewParserError(msg, literal)
			}

			// TODO
			rule.MaxRetweets = num
		default:
			msg := fmt.Sprintf("Invalid operator for \"reweets\": %s", op.val)
			return nil, NewParserError(msg, op)
		}
	default:
		msg := fmt.Sprintf("Invalid identifier: %s", ident.val)
		return nil, NewParserError(msg, op)
	}

	node := &ParseNode{
		kind:     "cond",
		rule:     rule,
		op:       op.kind,
		children: nil,
	}

	return node, nil
}

func (parser *Parser) ident() (*Token, error) {
	token, err := parser.match("IDENT")
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (parser *Parser) logical() (*Token, error) {
	var token *Token
	var err error

	switch parser.currToken.kind {
	case "AND", "OR":
		token, err = parser.match(parser.currToken.kind)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Invalid logical operator")
	}

	return token, nil
}

func (parser *Parser) op() (*Token, error) {
	var token *Token
	var err error

	switch parser.currToken.kind {
	case "GTE", "GT", "LTE", "LT", "EQ", "NEQ", "IN", "NOTIN":
		token, err = parser.match(parser.currToken.kind)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Invalid comparison operator")
	}

	return token, nil
}

func (parser *Parser) literal() (*Token, error) {
	var token *Token
	var err error

	switch parser.currToken.kind {
	case "STRING", "NUMBER", "AGE", "TIME":
		token, err = parser.match(parser.currToken.kind)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Invalid literal")
	}

	return token, nil
}

func PrintParsedRule(currNode *ParseNode, depth int) {
	if currNode.numChildren == 0 {
		return
	}

	for _, node := range currNode.children {
		fmt.Printf("depth = %d, %s\n", depth, node)
		PrintParsedRule(node, depth+1)
	}
}

// Reset this Parser to a clean state with the provided input
func (parser *Parser) Reset(input string) {
	parser.lexer = NewLexer(TOKENS, input)
	parser.rule = NewParsedRule(input)
}

// Entry point for parser
func (parser *Parser) Parse() (*ParsedRule, error) {
	token, err := parser.lexer.NextToken()
	if err != nil {
		return nil, err
	}

	parser.currToken = token

	_, err1 := parser.expr(parser.rule.root)

	return parser.rule, err1
}
