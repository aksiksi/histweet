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
		{"age > 3m && (likes < 100 && likes == 34)", 6},
		{"(age > 3m && (likes < 100 && likes >= 34)) || text !~ \"xyz\"", 9},
		{`((text !~ "hey!") && (likes == 5) && (likes == 3)) || ( likes == 9)`, 12},
		{`((text !~ "hey!") && (likes == 5)) || created < 10-May-2020 || likes == 9`, 10},
	}

	for _, input := range inputs {
		t.Run(input.input, func(t *testing.T) {
			rule, err := Parse(input.input)
			if err != nil {
				t.Errorf("Failed to parse: %s", err)
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
