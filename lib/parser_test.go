package histweet

import (
	"testing"
	"time"
)

func TestParser(t *testing.T) {
	// Ensures that parser can successfully parse some known inputs, and
	// returns the correct number of parse nodes
	var inputs = []struct {
		input    string
		numNodes int
	}{
		// Valid
		{"likes == 34 && retweets == 2", 3},
		{"likes != 34 && retweets != 2", 3},
		{"likes >= 34 && retweets >= 2", 3},
		{"likes <= 34 && retweets <= 2", 3},
		{"likes > 34 && retweets > 2", 3},
		{"likes < 34 && retweets < 2", 3},
		{"age > 3m && likes < 100", 3},
		{"(age > 3m && likes >= 34) || text !~ \"xyz\"", 6},
		{"(age > 3m && (retweets < 100 && likes >= 34)) || text !~ \"xyz\"", 9},
		{"(age < 5y3m2d && (likes < 100 && retweets != 34)) || text !~ \"xyz\"", 9},
		{`((text !~ "hey!") && (likes == 5) && (likes == 3)) || ( likes == 9)`, 12},
		{`((text !~ "hey!") && (likes == 5)) || created < 10-May-2020 || likes == 9`, 10},
		{`((text !~ "hey!") && (likes == 5)) || created > 10-May-2020 || likes == 9`, 10},

		// Invalid literals (from left to right)
		{`created > "xyz"`, -1},
		{"age > 123", -1},
		{"likes > 3m", -1},
		{"retweets <= 3m", -1},
		{`((text !~ 666) && (likes == 5)) || created < 10-May-2020 || likes == 9`, -1},
		{`((text !~ "hey!") && (likes == x)) || created < 10-May-2020 || likes == 9`, -1},
		{`((text !~ "hey!") && (likes == 5)) || created < 10-Potato-2020 || likes == 9`, -1},

		// Invalid identifiers
		{`hummus !~ "hey!" && likes == 5`, -1},
		{`text !~ "hey!" || hates == 5`, -1},
		{"potatoes > 10", -1},

		// Invalid operators
		{"age - 3m && (likes < 100 && likes == 34)", -1},
		{"age > 3m && (likes . 100 && likes == 34)", -1},
		{"retweets !~ 3m", -1},
		{"retweets !~ 10", -1},
		{`text < "abcd"`, -1},
		{`created ~ 10-May-2020`, -1},

		// Unbalanced parens
		{"(age > 3m && likes >= 34 || text !~ \"xyz\"", -1},
		{"age > 3m && likes >= 34) || text !~ \"xyz\"", -1},
	}

	for _, input := range inputs {
		t.Run(input.input, func(t *testing.T) {
			rule, err := Parse(input.input)
			if err != nil {
				if input.numNodes == -1 {
					t.Logf("Invalid rule detected -- %s", err)
					return
				}

				t.Errorf("Failed to parse rule: %s", err)
			}

			numNodes := rule.numNodes
			expected := input.numNodes

			if numNodes != expected {
				t.Errorf("Parsed count %d != expected %d\n", numNodes, expected)
			}

			s := rule.ToString()
			if s == "" {
				t.Errorf("No string returned\n")
			}

			// Evaluate a random Tweet on the parsed rule
			tweet := Tweet{
				Text:     "abc",
				NumLikes: 10,
			}

			rule.Eval(&tweet)
		})
	}
}

func BenchmarkParser(b *testing.B) {
	input := `((text !~ "hey!") && (likes == 5)) || created < 10-May-2020 || likes == 9`
	parser := NewParser(input)

	for i := 0; i < b.N; i++ {
		rule, err := parser.Parse()
		if err != nil {
			panic(err)
		}

		if rule == nil {
			panic("Rule is invalid")
		}

		parser.Reset(input)
	}
}

func BenchmarkEval(b *testing.B) {
	input := `((text !~ "hey!") && (likes == 5)) || created < 10-May-2020 || likes == 9`
	rule, _ := Parse(input)
	tweet := Tweet{
		Text:      "abc hey 123",
		NumLikes:  10,
		CreatedAt: time.Date(2020, 5, 11, 0, 0, 0, 0, time.UTC),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rule.Eval(&tweet)
	}
}
