package histweet

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const (
	timeLayout = "02-Jan-2006"
)

// TOKENS for terminals of the Twitter rule parser grammar
var TOKENS = map[tokenKind]string{
	tokenIdent:  "[a-zA-Z_]+",
	tokenNumber: "[0-9]+",
	tokenString: `"[^\"]*"`,
	tokenAge:    `^\s*([0-9]+[ymd])?([0-9]+[ymd])?([0-9]+[ymd])`,
	tokenTime:   `\d\d-\w\w\w-\d\d\d\d`,
	tokenLparen: `\(`,
	tokenRparen: `\)`,
	tokenOr:     `\|\|`,
	tokenAnd:    "&&",
	tokenGte:    ">=",
	tokenGt:     ">",
	tokenLte:    "<=",
	tokenLt:     "<",
	tokenEq:     "==",
	tokenNeq:    "!=",
	tokenIn:     "~",
	tokenNotIn:  "!~",
}

// ParseNode represents a single node in the parse tree.
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
	op          tokenKind
	rule        *RuleTweet
	children    []*ParseNode
	numChildren int
}

func (node *ParseNode) String() string {
	return fmt.Sprintf("Kind: %s, Op: %d, NumChildren: %d, Rule: %+v",
		node.kind, node.op, node.numChildren, node.rule)
}

// ParsedRule represents a single parsed Rule as a tree of ParseNodes.
type ParsedRule struct {
	root     *ParseNode
	input    string
	numNodes int
}

func newParsedRule(input string) *ParsedRule {
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
		case tokenAnd:
			return left && right
		case tokenOr:
			return left || right
		default:
			panic(fmt.Sprintf("Unexpected logical op: %d\n", node.op))
		}
	} else {
		// TODO: Does this make sense?
		return evalInternal(tweet, node.children[0])
	}
}

// IsMatch walks the parse tree and evaluates each condition against
// the given Tweet. Returns true if the Tweet matches all of the rules.
func (rule *ParsedRule) IsMatch(tweet *Tweet) bool {
	root := rule.root
	return evalInternal(tweet, root)
}

// Parser is a simple parser for tweet deletion rule strings.
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
	lexer *lexer

	// Pointer to the current token
	currToken *token

	// Tree of parse nodes
	rule *ParsedRule
}

// ParserError represents errors hit during rule parsing
type ParserError struct {
	msg   string
	token *token
}

func (err *ParserError) Error() string {
	return fmt.Sprintf("Parser Error: %s (col %d)", err.msg, err.token.pos)
}

func newParserError(msg string, token *token) *ParserError {
	return &ParserError{
		msg:   msg,
		token: token,
	}
}

// Verifies that current token is of the specified `kind`,
// returns it, and reads in the next token
func (parser *Parser) match(kind tokenKind) (*token, error) {
	currToken := parser.currToken

	if currToken.kind != kind {
		return nil, fmt.Errorf("Failed match for kind = %d", kind)
	}

	token, err := parser.lexer.nextToken()
	if err != nil {
		return nil, err
	}

	parser.currToken = token

	return currToken, nil
}

func (parser *Parser) expr(parent *ParseNode) (*ParseNode, error) {
	var token *token
	var err error
	var node, logicalNode *ParseNode

	if parser.currToken.kind == tokenLparen {
		_, err = parser.match(tokenLparen)
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

		_, err = parser.match(tokenRparen)
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
		if literal.kind != tokenAge {
			msg := fmt.Sprintf("Invalid literal for \"age\": %s", literal.val)
			return nil, newParserError(msg, literal)
		}

		time, err3 := convertAgeToTime(literal.val)
		if err3 != nil {
			msg := fmt.Sprintf("Invalid format for \"age\": %s", literal.val)
			return nil, newParserError(msg, literal)
		}

		switch op.kind {
		case tokenGt, tokenGte:
			rule.Before = time
		case tokenLt, tokenLte:
			rule.After = time
		default:
			msg := fmt.Sprintf("Invalid operator for \"age\": %s", op.val)
			return nil, newParserError(msg, op)
		}
	case "text":
		switch op.kind {
		case tokenIn, tokenNotIn:
			rule.Match = regexp.MustCompile(literal.val)
		default:
			msg := fmt.Sprintf("Invalid operator for \"text\": %s", op.val)
			return nil, newParserError(msg, op)
		}
	case "created":
		if literal.kind != tokenTime {
			msg := fmt.Sprintf("Invalid literal for \"created\": %s", literal.val)
			return nil, newParserError(msg, literal)
		}

		time, err4 := time.Parse(timeLayout, literal.val)
		if err4 != nil {
			msg := fmt.Sprintf("Invalid format for time for \"created\": %s", literal.val)
			return nil, newParserError(msg, literal)
		}

		switch op.kind {
		case tokenGt, tokenGte:
			rule.Before = time
		case tokenLt, tokenLte:
			rule.After = time
		default:
			msg := fmt.Sprintf("Invalid operator for \"created\": %s", op.val)
			return nil, newParserError(msg, op)
		}
	case "likes":
		switch op.kind {
		case tokenLt, tokenLte, tokenGt, tokenGte, tokenEq, tokenNeq:
			num, err := strconv.Atoi(literal.val)
			if err != nil {
				msg := fmt.Sprintf("Invalid number for \"likes\": %s", literal.val)
				return nil, newParserError(msg, literal)
			}

			// TODO
			rule.MaxLikes = num
		default:
			msg := fmt.Sprintf("Invalid operator for \"likes\": %s", op.val)
			return nil, newParserError(msg, op)
		}
	case "reweets":
		switch op.kind {
		case tokenLt, tokenLte, tokenGt, tokenGte, tokenEq, tokenNeq:
			num, err := strconv.Atoi(literal.val)
			if err != nil {
				msg := fmt.Sprintf("Invalid number for \"reweets\": %s", literal.val)
				return nil, newParserError(msg, literal)
			}

			// TODO
			rule.MaxRetweets = num
		default:
			msg := fmt.Sprintf("Invalid operator for \"reweets\": %s", op.val)
			return nil, newParserError(msg, op)
		}
	default:
		msg := fmt.Sprintf("Invalid identifier: %s", ident.val)
		return nil, newParserError(msg, op)
	}

	node := &ParseNode{
		kind:     "cond",
		rule:     rule,
		op:       op.kind,
		children: nil,
	}

	return node, nil
}

func (parser *Parser) ident() (*token, error) {
	token, err := parser.match(tokenIdent)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (parser *Parser) logical() (*token, error) {
	var token *token
	var err error

	switch parser.currToken.kind {
	case tokenAnd, tokenOr:
		token, err = parser.match(parser.currToken.kind)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Invalid logical operator")
	}

	return token, nil
}

func (parser *Parser) op() (*token, error) {
	var token *token
	var err error

	switch parser.currToken.kind {
	case tokenLt, tokenLte, tokenGt, tokenGte, tokenEq, tokenNeq, tokenIn, tokenNotIn:
		token, err = parser.match(parser.currToken.kind)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Invalid comparison operator")
	}

	return token, nil
}

func (parser *Parser) literal() (*token, error) {
	var token *token
	var err error

	switch parser.currToken.kind {
	case tokenString, tokenNumber, tokenAge, tokenTime:
		token, err = parser.match(parser.currToken.kind)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Invalid literal")
	}

	return token, nil
}

// PrintParsedRule is a helper that prints out a tree of ParseNodes
func PrintParsedRule(currNode *ParseNode, depth int) {
	if currNode.numChildren == 0 {
		return
	}

	for _, node := range currNode.children {
		fmt.Printf("depth = %d, %s\n", depth, node)
		PrintParsedRule(node, depth+1)
	}
}

// NewParser builds a new Parser from the input
func NewParser(input string) *Parser {
	lexer := newLexer(TOKENS, input)

	parser := &Parser{
		lexer: lexer,
		rule:  newParsedRule(input),
	}

	return parser
}

// Reset this Parser to a clean state with the provided input
func (parser *Parser) Reset(input string) {
	parser.lexer = newLexer(TOKENS, input)
	parser.rule = newParsedRule(input)
}

// Parse is the entry point for parser
func (parser *Parser) Parse() (*ParsedRule, error) {
	token, err := parser.lexer.nextToken()
	if err != nil {
		return nil, err
	}

	parser.currToken = token

	_, err1 := parser.expr(parser.rule.root)

	return parser.rule, err1
}
