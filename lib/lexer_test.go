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

		// Invalid tokens
		"age > 3m ** (likes < 100 && likes == 34)": {
			token{kind: tokenIdent, val: "age"},
			token{kind: tokenGt, val: ">"},
			token{kind: tokenAge, val: "3m"},
			token{kind: tokenEOF, val: "**"},
		},
		"-++": {
			token{kind: tokenEOF, val: "-"},
			token{kind: tokenEOF, val: "+"},
			token{kind: tokenEOF, val: "+"},
		},
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			lexer := newLexer(Tokens, input)

			for i := 0; i < len(expected); i++ {
				currExpected := &expected[i]

				token, err := lexer.nextToken()
				if err != nil {
					if currExpected.kind == tokenEOF {
						// Failed as expected
						t.Log("Found invalid token: ", currExpected)
						continue
					} else {
						t.Errorf("Error: %s", err)
					}
				}

				if token.kind != currExpected.kind {
					t.Errorf("Invalid kind - found: %d != expected: %d",
						token.kind, currExpected.kind)
				}

				if token.val != currExpected.val {
					t.Errorf("Invalid value - found: %s != expected: %s",
						token.val, currExpected.val)
				}
			}

			lexer.Reset()
			if lexer.pos != 0 || lexer.numTokens != 0 {
				t.Errorf("Lexer not reset properly")
			}
		})
	}
}

// Dummy test to trigger ToString() for lexer token kinds
func TestLexerTokenKinds(t *testing.T) {
	kinds := []tokenKind{
		tokenIdent,
		tokenNumber,
		tokenString,
		tokenAge,
		tokenTime,
		tokenLparen,
		tokenRparen,
		tokenOr,
		tokenAnd,
		tokenGte,
		tokenGt,
		tokenLte,
		tokenLt,
		tokenEq,
		tokenNeq,
		tokenIn,
		tokenNotIn,
		tokenEOF,
		9999,
	}

	for _, kind := range kinds {
		kind.ToString()
	}
}
