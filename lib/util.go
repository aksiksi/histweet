package histweet

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Helper function to check for unbalanced parens in a given string
func checkUnbalancedParens(input string) error {
	var parenStack []int

	for i, c := range input {
		l := len(parenStack)

		if c == '(' {
			parenStack = append(parenStack, i)
		} else if c == ')' {
			if l == 0 {
				return fmt.Errorf(`Unbalanced ")" at pos %d`, i)
			}

			parenStack = parenStack[:len(parenStack)-1]
		}
	}

	if len(parenStack) > 0 {
		l := len(parenStack)
		return fmt.Errorf(`Unbalanced "(" at pos %d`, parenStack[l-1])
	}

	return nil
}

func convertAgeToTime(age string) (time.Time, error) {
	var days int
	var months int
	var years int

	agePat := regexp.MustCompile(`(\d+y)?(\d+m)?(\d+d)?`)

	matches := agePat.FindStringSubmatch(age)
	if matches == nil {
		return time.Time{}, fmt.Errorf("Invalid age string provided: %s", age)
	}

	for _, match := range matches[1:] {
		if match == "" {
			continue
		}

		val, err := strconv.ParseInt(match[:len(match)-1], 10, 32)
		if err != nil {
			return time.Time{}, err
		}

		// The last character of this match must be one of: d, m, or y
		switch match[len(match)-1] {
		case 'd':
			days = int(val)
		case 'm':
			months = int(val)
		case 'y':
			years = int(val)
		default:
			return time.Time{}, fmt.Errorf("Invalid age string provided: must only contain \"d\", \"m\", or \"y\"")
		}
	}

	// This is how you go back in time
	now := time.Now().UTC()
	target := now.AddDate(-years, -months, -days)

	return target, nil
}
