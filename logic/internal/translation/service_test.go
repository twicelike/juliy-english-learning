package translation

import (
	"reflect"
	"testing"
)

func TestExtractJSONCandidates(t *testing.T) {
	tests := []struct {
		name      string
		response  string
		expected  []string
		expectErr bool
	}{
		{
			name:      "Valid JSON array",
			response:  `["tiền lương", "tiền công"]`,
			expected:  []string{"tiền lương", "tiền công"},
			expectErr: false,
		},
		{
			name:      "JSON array wrapped in markdown",
			response:  "```json\n[\"ngân hàng\", \"bờ sông\"]\n```",
			expected:  []string{"ngân hàng", "bờ sông"},
			expectErr: false,
		},
		{
			name:      "JSON with garbage text around it",
			response:  "Here is your translation:\n[\"cà chua\"]\nHope this helps!",
			expected:  []string{"cà chua"},
			expectErr: false,
		},
		{
			name:      "Invalid JSON",
			response:  `{"translation": "invalid"}`,
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "Duplicates and empty strings with whitespace",
			response:  `["bank", " bank ", "", "BANK", "river"]`,
			expected:  []string{"bank", "river"}, // duplicate ignored case-insensitively, empty ignored
			expectErr: false,
		},
		{
			name:      "More than 5 candidates",
			response:  `["one", "two", "three", "four", "five", "six"]`,
			expected:  []string{"one", "two", "three", "four", "five"}, // cut off at 5
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractJSONCandidates(tt.response)
			if (err != nil) != tt.expectErr {
				t.Errorf("extractJSONCandidates() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("extractJSONCandidates() got = %v, expected %v", got, tt.expected)
			}
		})
	}
}
