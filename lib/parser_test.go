package histweet

import (
	"testing"
)

func TestLexer(t *testing.T) {
	var tests = map[string][]Token{
		"(size < 100 && weight == 34)": {
			Token{kind: tokenLparen, val: "("},
			Token{kind: tokenIdent, val: "size"},
			Token{kind: tokenLt, val: "<"},
			Token{kind: tokenNumber, val: "100"},
			Token{kind: tokenAnd, val: "&&"},
			Token{kind: tokenIdent, val: "weight"},
			Token{kind: tokenEq, val: "=="},
			Token{kind: tokenNumber, val: "34"},
			Token{kind: tokenRparen, val: ")"},
		},
		"(size < 100 && created == 10-May-2020)": {
			Token{kind: tokenLparen, val: "("},
			Token{kind: tokenIdent, val: "size"},
			Token{kind: tokenLt, val: "<"},
			Token{kind: tokenNumber, val: "100"},
			Token{kind: tokenAnd, val: "&&"},
			Token{kind: tokenIdent, val: "created"},
			Token{kind: tokenEq, val: "=="},
			Token{kind: tokenTime, val: "10-May-2020"},
			Token{kind: tokenRparen, val: ")"},
		},
		"age > 3m && (size < 100 && weight == 34)": {
			Token{kind: tokenIdent, val: "age"},
			Token{kind: tokenGt, val: ">"},
			Token{kind: tokenAge, val: "3m"},
			Token{kind: tokenAnd, val: "&&"},
			Token{kind: tokenLparen, val: "("},
			Token{kind: tokenIdent, val: "size"},
			Token{kind: tokenLt, val: "<"},
			Token{kind: tokenNumber, val: "100"},
			Token{kind: tokenAnd, val: "&&"},
			Token{kind: tokenIdent, val: "weight"},
			Token{kind: tokenEq, val: "=="},
			Token{kind: tokenNumber, val: "34"},
			Token{kind: tokenRparen, val: ")"},
		},
		"(age > 3m && (size < 100 && weight >= 34)) || text !~ \"xyz\"": {
			Token{kind: tokenLparen, val: "("},
			Token{kind: tokenIdent, val: "age"},
			Token{kind: tokenGt, val: ">"},
			Token{kind: tokenAge, val: "3m"},
			Token{kind: tokenAnd, val: "&&"},
			Token{kind: tokenLparen, val: "("},
			Token{kind: tokenIdent, val: "size"},
			Token{kind: tokenLt, val: "<"},
			Token{kind: tokenNumber, val: "100"},
			Token{kind: tokenAnd, val: "&&"},
			Token{kind: tokenIdent, val: "weight"},
			Token{kind: tokenGte, val: ">="},
			Token{kind: tokenNumber, val: "34"},
			Token{kind: tokenRparen, val: ")"},
			Token{kind: tokenRparen, val: ")"},
			Token{kind: tokenOr, val: "||"},
			Token{kind: tokenIdent, val: "text"},
			Token{kind: tokenNotIn, val: "!~"},
			Token{kind: tokenString, val: `"xyz"`},
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
					t.Errorf("Invalid kind - found: %d != expected: %d",
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
