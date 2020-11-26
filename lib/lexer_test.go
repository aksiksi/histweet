package histweet

import (
	"testing"
)

func TestLexer(t *testing.T) {
	var tests = map[string][]token{
		"(size < 100 && weight == 34)": {
			token{kind: tokenLparen, val: "("},
			token{kind: tokenIdent, val: "size"},
			token{kind: tokenLt, val: "<"},
			token{kind: tokenNumber, val: "100"},
			token{kind: tokenAnd, val: "&&"},
			token{kind: tokenIdent, val: "weight"},
			token{kind: tokenEq, val: "=="},
			token{kind: tokenNumber, val: "34"},
			token{kind: tokenRparen, val: ")"},
		},
		"(size < 100 && created == 10-May-2020)": {
			token{kind: tokenLparen, val: "("},
			token{kind: tokenIdent, val: "size"},
			token{kind: tokenLt, val: "<"},
			token{kind: tokenNumber, val: "100"},
			token{kind: tokenAnd, val: "&&"},
			token{kind: tokenIdent, val: "created"},
			token{kind: tokenEq, val: "=="},
			token{kind: tokenTime, val: "10-May-2020"},
			token{kind: tokenRparen, val: ")"},
		},
		"age > 3m && (size < 100 && weight == 34)": {
			token{kind: tokenIdent, val: "age"},
			token{kind: tokenGt, val: ">"},
			token{kind: tokenAge, val: "3m"},
			token{kind: tokenAnd, val: "&&"},
			token{kind: tokenLparen, val: "("},
			token{kind: tokenIdent, val: "size"},
			token{kind: tokenLt, val: "<"},
			token{kind: tokenNumber, val: "100"},
			token{kind: tokenAnd, val: "&&"},
			token{kind: tokenIdent, val: "weight"},
			token{kind: tokenEq, val: "=="},
			token{kind: tokenNumber, val: "34"},
			token{kind: tokenRparen, val: ")"},
		},
		"(age > 3m && (size < 100 && weight >= 34)) || text !~ \"xyz\"": {
			token{kind: tokenLparen, val: "("},
			token{kind: tokenIdent, val: "age"},
			token{kind: tokenGt, val: ">"},
			token{kind: tokenAge, val: "3m"},
			token{kind: tokenAnd, val: "&&"},
			token{kind: tokenLparen, val: "("},
			token{kind: tokenIdent, val: "size"},
			token{kind: tokenLt, val: "<"},
			token{kind: tokenNumber, val: "100"},
			token{kind: tokenAnd, val: "&&"},
			token{kind: tokenIdent, val: "weight"},
			token{kind: tokenGte, val: ">="},
			token{kind: tokenNumber, val: "34"},
			token{kind: tokenRparen, val: ")"},
			token{kind: tokenRparen, val: ")"},
			token{kind: tokenOr, val: "||"},
			token{kind: tokenIdent, val: "text"},
			token{kind: tokenNotIn, val: "!~"},
			token{kind: tokenString, val: `"xyz"`},
		},
	}

	for input, expected := range tests {
		lexer := newLexer(TOKENS, input)

		t.Run(input, func(t *testing.T) {
			for i := 0; i < len(expected); i++ {
				token, err := lexer.nextToken()
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
