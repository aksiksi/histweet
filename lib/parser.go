package histweet

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	timeLayout = "02-Jan-2006"
)

// Tokens for terminals of the Twitter rule parser grammar
//
// All token regular expressions _must_ start with ^ to ensure
// that the match is computed from the current position in the
// stream.
var Tokens = map[tokenKind]string{
	tokenIdent:  "^[a-zA-Z_]+",
	tokenNumber: "^[0-9]+",
	tokenString: `^"[^\"]*"`,
	tokenAge:    `^\s*([0-9]+[ymd])?([0-9]+[ymd])?([0-9]+[ymd])`,
	tokenTime:   `^\d\d-\w\w\w-\d\d\d\d`,
	tokenLparen: `^\(`,
	tokenRparen: `^\)`,
	tokenOr:     `^\|\|`,
	tokenAnd:    "^&&",
	tokenGte:    "^>=",
	tokenGt:     "^>",
	tokenLte:    "^<=",
	tokenLt:     "^<",
	tokenEq:     "^==",
	tokenNeq:    "^!=",
	tokenIn:     "^~",
	tokenNotIn:  "^!~",
}

type nodeKind int

// Types of parser nodes
const (
	nodeCond nodeKind = iota
	nodeLogical
)

// parseNode represents a single node in the parse tree.
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
type parseNode struct {
	kind  nodeKind
	op    tokenKind
	rule  *RuleTweet
	left  *parseNode
	right *parseNode
}

func (node *parseNode) String() string {
	return fmt.Sprintf("Kind: %d, Op: %d, Rule: %+v", node.kind, node.op, node.rule)
}

// ParsedRule represents a single parsed Rule as a tree of parseNodes.
type ParsedRule struct {
	root     *parseNode
	numNodes int
}

func evalInternal(tweet *Tweet, node *parseNode) bool {
	switch node.kind {
	case nodeCond:
		return tweet.IsMatch(node.rule)
	case nodeLogical:
		left := evalInternal(tweet, node.left)
		right := evalInternal(tweet, node.right)

		switch node.op {
		case tokenAnd:
			return left && right
		case tokenOr:
			return left || right
		default:
			panic(fmt.Sprintf("Unexpected logical op: %d\n", node.op))
		}
	default:
		panic(fmt.Sprintf("Unexpected node type: %d", node.kind))
	}
}

// Eval walks the parse tree and evaluates each condition against
// the given Tweet. Returns true if the Tweet matches all of the rules.
func (rule *ParsedRule) Eval(tweet *Tweet) bool {
	return evalInternal(tweet, rule.root)
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
// Expr    <-  ( Expr ) | Cond [Logical Expr]?
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
	msg  string
	pos  int
	kind tokenKind
	val  string
	// TODO: Add line
}

func (err *ParserError) Error() string {
	return fmt.Sprintf("%s: \"%s\" (%s) (at col %d)", err.msg, err.val, err.kind.ToString(), err.pos)
}

func newParserError(msg string, token *token) *ParserError {
	return &ParserError{
		msg:  msg,
		pos:  token.pos,
		kind: token.kind,
		val:  token.val,
	}
}

// Verifies that current token is of the specified `kind`,
// returns it, and reads in the next token
func (parser *Parser) match(kind tokenKind) (*token, error) {
	currToken := parser.currToken

	// If the current token is not a match, return the token for
	// error reporting purposes. Do not consume the token.
	if currToken.kind != kind {
		return currToken, fmt.Errorf(`Unexpected token - found: "%s", expected: "%s"`,
			currToken.kind.ToString(), kind.ToString())
	}

	token, err := parser.lexer.nextToken()
	if err != nil {
		return nil, err
	}

	parser.currToken = token

	return currToken, nil
}

func (parser *Parser) expr() (*parseNode, error) {
	var node *parseNode
	var err error

	for {
		token := parser.currToken

		// TODO(aksiksi): Handle the case of a non-logical expression that follows
		// an expression or cond
		switch token.kind {
		// Nested expression
		case tokenLparen:
			_, err = parser.match(tokenLparen)
			if err != nil {
				return nil, err
			}

			// Parse the internal expression and return the resulting node
			node, err = parser.expr()
			if err != nil {
				return nil, err
			}

			token, err = parser.match(tokenRparen)
			if err != nil {
				return nil, err
			}
		// Conditional expression
		case tokenIdent:
			node, err = parser.cond()
			if err != nil {
				return nil, err
			}
		// Logical/binary expression
		case tokenAnd, tokenOr:
			// Logical expresion with no preceding expression is invalid
			if node == nil {
				return nil, fmt.Errorf("Unexpected logical operator at %d: %s", token.pos, token.kind.ToString())
			}

			op, err := parser.logical()
			if err != nil {
				return nil, err
			}

			newNode, err := parser.expr()
			if err != nil {
				return nil, err
			}

			node = &parseNode{
				kind:  nodeLogical,
				op:    op.kind,
				rule:  nil,
				left:  node,
				right: newNode,
			}
		default:
			return node, nil
		}

		parser.rule.numNodes++
	}
}

func (parser *Parser) cond() (*parseNode, error) {
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
			return nil, newParserError("Invalid literal for \"age\"", literal)
		}

		time, err3 := convertAgeToTime(literal.val)
		if err3 != nil {
			return nil, newParserError("Invalid format for \"age\"", literal)
		}

		switch op.kind {
		case tokenGt, tokenGte:
			rule.Before = time
		case tokenLt, tokenLte:
			rule.After = time
		default:
			return nil, newParserError("Invalid operator for \"age\"", op)
		}
	case "text":
		if literal.kind != tokenString {
			return nil, newParserError("Invalid literal for \"text\"", literal)
		}

		switch op.kind {
		case tokenIn, tokenNotIn:
			// Gotcha: the literal contains quotes - remove them before building the regexp
			pat := strings.Replace(literal.val, "\"", "", 2)

			rule.Match = regexp.MustCompile(pat)
		default:
			return nil, newParserError("Invalid operator for \"text\"", op)
		}
	case "created":
		if literal.kind != tokenTime {
			return nil, newParserError("Invalid literal for \"created\"", literal)
		}

		time, err4 := time.Parse(timeLayout, literal.val)
		if err4 != nil {
			return nil, newParserError("Invalid time format for \"created\"", literal)
		}

		switch op.kind {
		case tokenGt, tokenGte:
			rule.Before = time
		case tokenLt, tokenLte:
			rule.After = time
		default:
			return nil, newParserError("Invalid operator for \"created\"", op)
		}
	case "likes":
		if literal.kind != tokenNumber {
			return nil, newParserError("Invalid literal for \"likes\"", literal)
		}

		num, err := strconv.Atoi(literal.val)
		if err != nil {
			return nil, newParserError("Invalid number for \"likes\"", literal)
		}

		rule.Likes = num

		switch op.kind {
		case tokenGt:
			rule.LikesComparator = comparatorGt
		case tokenGte:
			rule.LikesComparator = comparatorGte
		case tokenLt:
			rule.LikesComparator = comparatorLt
		case tokenLte:
			rule.LikesComparator = comparatorLte
		case tokenEq:
			rule.LikesComparator = comparatorEq
		case tokenNeq:
			rule.LikesComparator = comparatorNeq
		default:
			return nil, newParserError("Invalid operator for \"likes\"", op)
		}
	case "retweets":
		if literal.kind != tokenNumber {
			return nil, newParserError("Invalid literal for \"retweets\"", literal)
		}

		num, err := strconv.Atoi(literal.val)
		if err != nil {
			return nil, newParserError("Invalid number for \"retweets\"", literal)
		}

		rule.Retweets = num

		switch op.kind {
		case tokenGt:
			rule.RetweetsComparator = comparatorGt
		case tokenGte:
			rule.RetweetsComparator = comparatorGte
		case tokenLt:
			rule.RetweetsComparator = comparatorLt
		case tokenLte:
			rule.RetweetsComparator = comparatorLte
		case tokenEq:
			rule.RetweetsComparator = comparatorEq
		case tokenNeq:
			rule.RetweetsComparator = comparatorNeq
		default:
			return nil, newParserError("Invalid operator for \"retweets\"", op)
		}
	default:
		return nil, newParserError("Invalid identifier", ident)
	}

	node := &parseNode{
		kind: nodeCond,
		rule: rule,
		op:   op.kind,
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
	token := parser.currToken

	switch token.kind {
	case tokenAnd, tokenOr:
		token, err := parser.match(token.kind)
		if err != nil {
			return nil, err
		}

		return token, nil
	default:
		return nil, newParserError("Invalid operator for logical expression", token)
	}
}

func (parser *Parser) op() (*token, error) {
	token := parser.currToken

	switch token.kind {
	case tokenLt, tokenLte, tokenGt, tokenGte, tokenEq, tokenNeq, tokenIn, tokenNotIn:
		token, err := parser.match(parser.currToken.kind)
		if err != nil {
			return nil, err
		}

		return token, nil
	default:
		return nil, newParserError("Invalid comparison operator", token)
	}
}

func (parser *Parser) literal() (*token, error) {
	token := parser.currToken

	switch token.kind {
	case tokenString, tokenNumber, tokenAge, tokenTime:
		token, err := parser.match(token.kind)
		if err != nil {
			return nil, err
		}

		return token, nil
	default:
		return nil, newParserError("Invalid literal", token)
	}
}

func toStringHelper(p *parseNode, depth int, output *strings.Builder) {
	if p == nil {
		return
	}

	s := fmt.Sprintf("depth = %d, %s", depth, p)
	output.WriteString(s)

	toStringHelper(p.left, depth+1, output)
	toStringHelper(p.right, depth+1, output)
}

// ToString walks the parse tree and outputs it in string form
func (rule *ParsedRule) ToString() string {
	var output strings.Builder
	toStringHelper(rule.root, 0, &output)
	return output.String()
}

// NewParser builds a new Parser from the input
func NewParser(input string) *Parser {
	lexer := newLexer(Tokens, input)

	parser := &Parser{
		lexer: lexer,
		rule:  &ParsedRule{},
	}

	return parser
}

// Reset this Parser to a clean state with the provided input
func (parser *Parser) Reset(input string) {
	parser.lexer.Reset()
	parser.rule = &ParsedRule{}
}

// Parse is the entry point for parser
func (parser *Parser) Parse() (*ParsedRule, error) {
	// Prior to parsing a rule, check for unbalanced parens
	err := checkUnbalancedParens(parser.lexer.input)
	if err != nil {
		return nil, err
	}

	token, err := parser.lexer.nextToken()
	if err != nil {
		return nil, err
	}

	parser.currToken = token

	node, err := parser.expr()

	// Set the root to the returned root
	parser.rule.root = node

	return parser.rule, err
}

// Parse is the entry point to the rule parser infra.
// Users of the library should only be using this function.
func Parse(input string) (*ParsedRule, error) {
	parser := NewParser(input)

	rule, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	return rule, nil
}
