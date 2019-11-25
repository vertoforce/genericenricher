package enrichers

import "testing"

func TestStringSizeToInt(t *testing.T) {
	tests := []struct {
		input  string
		output uint64
	}{
		// Simple
		{"1b", 1},
		{"1kb", 1024},
		{"1mb", 1024 * 1024},
		{"1gb", 1024 * 1024 * 1024},
		{"1tb", 1024 * 1024 * 1024 * 1024},

		// With real numbers
		{"113b", 113},
		{"113kb", 113 * 1024},
		{"113mb", 113 * 1024 * 1024},
		{"113gb", 113 * 1024 * 1024 * 1024},
		{"113tb", 113 * 1024 * 1024 * 1024 * 1024},

		// Decimal
		{"5.5b", 5},
		{"5.1b", 5},
		{"5.9b", 5},
		{"5.5kb", 5632},
		{"5.1kb", 5222},
		{"5.9kb", 6041},

		// Unexpected
		{"abc", 0xFFFFFFFFFFFFFFFF},
		{"string", 0xFFFFFFFFFFFFFFFF},
		{"-1", 0xFFFFFFFFFFFFFFFF},
	}

	for i, test := range tests {
		if got := stringSizeToUint(test.input); got != test.output {
			t.Errorf("Test %d failed, wanted %d got %d", i, test.output, got)
		}
	}
}
