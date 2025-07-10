package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tiernacity/ratchet/internal/errors"
)

func TestParser_Parse(t *testing.T) {
	p := New()

	tests := []struct {
		name    string
		output  string
		want    float64
		wantErr bool
		errType string
	}{
		// Valid single numbers
		{
			name:   "simple integer",
			output: "42",
			want:   42.0,
		},
		{
			name:   "simple float",
			output: "3.14",
			want:   3.14,
		},
		{
			name:   "negative integer",
			output: "-5",
			want:   -5.0,
		},
		{
			name:   "negative float",
			output: "-2.71",
			want:   -2.71,
		},
		{
			name:   "scientific notation positive",
			output: "1e6",
			want:   1000000.0,
		},
		{
			name:   "scientific notation negative",
			output: "1.5e-3",
			want:   0.0015,
		},
		{
			name:   "uppercase scientific notation",
			output: "2E+5",
			want:   200000.0,
		},
		{
			name:   "zero",
			output: "0",
			want:   0.0,
		},
		{
			name:   "zero float",
			output: "0.0",
			want:   0.0,
		},

		// Numbers with surrounding text
		{
			name:   "number with prefix text",
			output: "Coverage: 85.5%",
			want:   85.5,
		},
		{
			name:   "number with suffix text",
			output: "42 errors found",
			want:   42.0,
		},
		{
			name:   "number with surrounding text",
			output: "Found 123 TODO comments in the codebase",
			want:   123.0,
		},
		{
			name:   "multiline with number",
			output: "Building project...\nTests passed: 95.2%\nDone.",
			want:   95.2,
		},

		// Error cases
		{
			name:    "empty string",
			output:  "",
			wantErr: true,
			errType: "parse",
		},
		{
			name:    "whitespace only",
			output:  "   \n\t  ",
			wantErr: true,
			errType: "parse",
		},
		{
			name:    "no numbers",
			output:  "hello world",
			wantErr: true,
			errType: "parse",
		},
		{
			name:    "multiple numbers",
			output:  "Found 42 errors and 13 warnings",
			wantErr: true,
			errType: "parse",
		},
		{
			name:    "percentage with text",
			output:  "Coverage increased from 80% to 85%",
			wantErr: true,
			errType: "parse",
		},

		// Edge cases
		{
			name:   "very large number",
			output: "999999999999",
			want:   999999999999.0,
		},
		{
			name:   "very small number",
			output: "0.000001",
			want:   0.000001,
		},
		{
			name:   "number with leading zeros",
			output: "0042",
			want:   42.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.Parse(tt.output)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType == "parse" {
					assert.True(t, errors.IsParseError(err), "expected ParseError, got %T", err)
				}
				return
			}
			
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractNumbers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []float64
	}{
		{
			name:  "single integer",
			input: "42",
			want:  []float64{42.0},
		},
		{
			name:  "single float",
			input: "3.14",
			want:  []float64{3.14},
		},
		{
			name:  "multiple numbers",
			input: "Found 42 errors and 13 warnings",
			want:  []float64{42.0, 13.0},
		},
		{
			name:  "scientific notation",
			input: "1e6 and 2.5e-3",
			want:  []float64{1000000.0, 0.0025},
		},
		{
			name:  "negative numbers",
			input: "-5 and -3.14",
			want:  []float64{-5.0, -3.14},
		},
		{
			name:  "no numbers",
			input: "hello world",
			want:  nil,
		},
		{
			name:  "mixed with text",
			input: "Version 1.2.3 has 99.9% compatibility",
			want:  []float64{1.2, 3.0, 99.9},
		},
		{
			name:  "percentages",
			input: "Coverage: 85.5% (was 80.2%)",
			want:  []float64{85.5, 80.2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractNumbers(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParser_ErrorMessages(t *testing.T) {
	p := New()

	tests := []struct {
		name       string
		output     string
		wantErrMsg string
	}{
		{
			name:       "empty output",
			output:     "",
			wantErrMsg: "invalid metric output '': output is empty",
		},
		{
			name:       "no numbers",
			output:     "hello world",
			wantErrMsg: "invalid metric output 'hello world': no numeric value found",
		},
		{
			name:       "multiple numbers",
			output:     "42 and 13",
			wantErrMsg: "invalid metric output '42 and 13': multiple numeric values found ([42 13]); output must contain exactly one number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.output)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErrMsg, err.Error())
		})
	}
}