package histweet

import (
	"testing"
)

func TestLexer(t *testing.T) {
	var tests = map[string][]Token{
		"(size < 100 && weight == 34)": {
			Token{kind: "LPAREN", val: "("},
			Token{kind: "IDENT", val: "size"},
			Token{kind: "LT", val: "<"},
			Token{kind: "NUMBER", val: "100"},
			Token{kind: "AND", val: "&&"},
			Token{kind: "IDENT", val: "weight"},
			Token{kind: "EQ", val: "=="},
			Token{kind: "NUMBER", val: "34"},
			Token{kind: "RPAREN", val: ")"},
		},
		"(size < 100 && created == 10-May-2020)": {
			Token{kind: "LPAREN", val: "("},
			Token{kind: "IDENT", val: "size"},
			Token{kind: "LT", val: "<"},
			Token{kind: "NUMBER", val: "100"},
			Token{kind: "AND", val: "&&"},
			Token{kind: "IDENT", val: "created"},
			Token{kind: "EQ", val: "=="},
			Token{kind: "TIME", val: "10-May-2020"},
			Token{kind: "RPAREN", val: ")"},
		},
		"age > 3m && (size < 100 && weight == 34)": {
			Token{kind: "IDENT", val: "age"},
			Token{kind: "GT", val: ">"},
			Token{kind: "AGE", val: "3m"},
			Token{kind: "AND", val: "&&"},
			Token{kind: "LPAREN", val: "("},
			Token{kind: "IDENT", val: "size"},
			Token{kind: "LT", val: "<"},
			Token{kind: "NUMBER", val: "100"},
			Token{kind: "AND", val: "&&"},
			Token{kind: "IDENT", val: "weight"},
			Token{kind: "EQ", val: "=="},
			Token{kind: "NUMBER", val: "34"},
			Token{kind: "RPAREN", val: ")"},
		},
		"(age > 3m && (size < 100 && weight >= 34)) || text !~ \"xyz\"": {
			Token{kind: "LPAREN", val: "("},
			Token{kind: "IDENT", val: "age"},
			Token{kind: "GT", val: ">"},
			Token{kind: "AGE", val: "3m"},
			Token{kind: "AND", val: "&&"},
			Token{kind: "LPAREN", val: "("},
			Token{kind: "IDENT", val: "size"},
			Token{kind: "LT", val: "<"},
			Token{kind: "NUMBER", val: "100"},
			Token{kind: "AND", val: "&&"},
			Token{kind: "IDENT", val: "weight"},
			Token{kind: "GTE", val: ">="},
			Token{kind: "NUMBER", val: "34"},
			Token{kind: "RPAREN", val: ")"},
			Token{kind: "RPAREN", val: ")"},
			Token{kind: "OR", val: "||"},
			Token{kind: "IDENT", val: "text"},
			Token{kind: "NOTIN", val: "!~"},
			Token{kind: "STRING", val: `"xyz"`},
		},
	}

	for input, expected := range tests {
		lexer := NewLexer(TOKENS, input)

		t.Run(input, func(t *testing.T) {
			for i := 0; i < len(expected); i++ {
				token, err := lexer.NextToken()
				if err != nil {
					t.Errorf("Error: %s", err)
				}

				currExpected := &expected[i]

				if token.kind != currExpected.kind {
					t.Errorf("Invalid kind - found: %s != expected: %s",
						token.kind, currExpected.kind)
				}

				if token.val != currExpected.val {
					t.Errorf("Invalid value - found: %s != expected: %s",
						token.val, currExpected.val)
				}
			}
		})
	}
}

func TestParser(t *testing.T) {
	// Ensures that parser can successfully parse some known inputs, and
	// returns the correct number of parse nodes
	var inputs = []struct {
		input    string
		numNodes int
	}{
		{"age > 3m && (likes < 100 && likes == 34)", 6},
		{"(age > 3m && (likes < 100 && likes >= 34)) || text !~ \"xyz\"", 9},
		{`((text !~ "hey!") && (likes == 5) && (likes == 3)) || ( likes == 9)`, 12},
		{`((text !~ "hey!") && (likes == 5)) || created < 10-May-2020 || likes == 9`, 10},
	}

	for _, input := range inputs {
		parser := NewParser(input.input)

		t.Run(input.input, func(t *testing.T) {
			_, err := parser.Parse()
			if err != nil {
				t.Errorf("Failed to parse: %s", err)
			}

			numNodes := parser.rule.numNodes
			expected := input.numNodes

			if numNodes != expected {
				t.Errorf("Parsed not count %d != expected %d\n", numNodes, expected)
			}
		})
	}
}
