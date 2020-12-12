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

func TestParserEval(t *testing.T) {
	// Checks that parser evaluates rules correctly
	var inputs = []struct {
		rule  string
		tweet Tweet
	}{
		{"likes == 34 && retweets == 2", Tweet{NumLikes: 34, NumRetweets: 2}},
		{"likes != 34 && retweets != 2", Tweet{NumLikes: 10, NumRetweets: 5}},
		{"likes != 34 || retweets == 2", Tweet{NumLikes: 34, NumRetweets: 2}},
		{"(likes != 34) || retweets == 2", Tweet{NumLikes: 34, NumRetweets: 2}},
		{"likes >= 34 && retweets >= 2", Tweet{NumLikes: 35, NumRetweets: 2}},
		{"likes <= 34 && retweets <= 2", Tweet{NumLikes: 33, NumRetweets: 2}},
		{"likes > 34 && retweets > 2", Tweet{NumLikes: 35, NumRetweets: 3}},
		{"likes < 34 && retweets < 2", Tweet{NumLikes: 33, NumRetweets: 1}},
		{"age > 3m && likes < 100", Tweet{
			CreatedAt: time.Now().AddDate(0, -3, -1),
			NumLikes:  99,
		}},
		{`text ~ "def" && retweets == 3`, Tweet{
			Text:        "def",
			NumRetweets: 3,
		}},
		{`((text !~ "abc") && (likes == 5)) || created < 10-May-2020 || likes == 9`, Tweet{
			Text:     "def",
			NumLikes: 5,
		}},
		{`((text !~ "abc") && (likes == 5)) || created < 10-May-2020 || likes == 9`, Tweet{
			Text:      "abc",
			NumLikes:  6,
			CreatedAt: time.Date(2020, 5, 11, 0, 0, 0, 0, time.UTC),
		}},
		{`((text !~ "abc") && (likes == 5)) || created < 10-May-2020 || likes == 9`, Tweet{
			NumLikes: 9,
		}},
	}

	for _, input := range inputs {
		t.Run(input.rule, func(t *testing.T) {
			rule, err := Parse(input.rule)
			if err != nil {
				t.Errorf("Failed to parse rule: %s", err)
			}

			isMatch := rule.Eval(&input.tweet)
			if !isMatch {
				t.Errorf("Failed to evaluate rule: %s, %v", input.rule, input.tweet)
			}
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
	input := `
		(((text !~ "hey!") && (likes == 5)) && text ~ "abc") || (created < 10-May-2020 || likes == 9) && (age > 1d || likes == 9)
	`
	rule, _ := Parse(input)
	tweet := Tweet{
		Text:      "abc hey 123",
		NumLikes:  9,
		CreatedAt: time.Date(2020, 5, 11, 0, 0, 0, 0, time.UTC),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rule.Eval(&tweet)
	}
}
