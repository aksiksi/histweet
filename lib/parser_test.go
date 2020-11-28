package histweet

import (
	"testing"
)

func TestParser(t *testing.T) {
	// Ensures that parser can successfully parse some known inputs, and
	// returns the correct number of parse nodes
	var inputs = []struct {
		input    string
		numNodes int
	}{
		// Valid
		{"age > 3m && (likes < 100 && likes == 34)", 6},
		{"(age > 3m && (likes < 100 && likes >= 34)) || text !~ \"xyz\"", 9},
		{`((text !~ "hey!") && (likes == 5) && (likes == 3)) || ( likes == 9)`, 12},
		{`((text !~ "hey!") && (likes == 5)) || created < 10-May-2020 || likes == 9`, 10},

		// Invalid literals (from left to right)
		{`((text !~ 666) && (likes == 5)) || created < 10-May-2020 || likes == 9`, -1},
		{`((text !~ "hey!") && (likes == x)) || created < 10-May-2020 || likes == 9`, -1},
		{`((text !~ "hey!") && (likes == 5)) || created < 10-Potato-2020 || likes == 9`, -1},
		{`((text !~ "hey!") && (likes == 5)) || created < 10-May-2020 || likes == abc`, -1},

		// Invalid identifiers
		{`hummus !~ "hey!" && likes == 5`, -1},
		{`text !~ "hey!" || hates == 5`, -1},

		// Invalid operators
		{"age - 3m && (likes < 100 && likes == 34)", -1},
		{"age > 3m ** (likes < 100 && likes == 34)", -1},
		{"age > 3m && (likes . 100 && likes == 34)", -1},
	}

	for _, input := range inputs {
		parser := NewParser(input.input)

		t.Run(input.input, func(t *testing.T) {
			parser.Reset(input.input)

			rule, err := parser.Parse()
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
				t.Errorf("Parsed not count %d != expected %d\n", numNodes, expected)
			}

			s := rule.ToString()
			if s == "" {
				t.Errorf("No string returned\n")
			}
		})
	}
}
