package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/tiernacity/ratchet/internal/errors"
)

// Parser interface for parsing metric output
type Parser interface {
	Parse(output string) (float64, error)
}

// parser implements the Parser interface
type parser struct{}

// New creates a new Parser instance
func New() Parser {
	return &parser{}
}

// Parse extracts a single numeric value from command output
func (p *parser) Parse(output string) (float64, error) {
	// Clean the output
	cleaned := strings.TrimSpace(output)
	if cleaned == "" {
		return 0, errors.WrapParseError(output, fmt.Errorf("output is empty"))
	}

	// Look for numeric patterns in the output
	numbers := extractNumbers(cleaned)

	switch len(numbers) {
	case 0:
		return 0, errors.WrapParseError(output, fmt.Errorf("no numeric value found"))
	case 1:
		return numbers[0], nil
	default:
		return 0, errors.WrapParseError(output, 
			fmt.Errorf("multiple numeric values found (%v); output must contain exactly one number", numbers))
	}
}

// extractNumbers finds all numeric values in the input string
func extractNumbers(input string) []float64 {
	// Regex to match integers and floating-point numbers (including scientific notation)
	// Matches: 42, -42, 3.14, -3.14, 1e10, -1.5e-10, 1E+5, etc.
	numberRegex := regexp.MustCompile(`-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?`)
	
	matches := numberRegex.FindAllString(input, -1)
	var numbers []float64
	
	for _, match := range matches {
		if num, err := strconv.ParseFloat(match, 64); err == nil {
			numbers = append(numbers, num)
		}
	}
	
	return numbers
}