package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseNumber parses a string output and returns a float64 number
func ParseNumber(output string) (float64, error) {
	// Trim whitespace
	output = strings.TrimSpace(output)

	if output == "" {
		return 0, fmt.Errorf("empty output")
	}

	// Try to parse as integer first
	if intVal, err := strconv.ParseInt(output, 10, 64); err == nil {
		return float64(intVal), nil
	}

	// Try to parse as float
	floatVal, err := strconv.ParseFloat(output, 64)
	if err != nil {
		return 0, fmt.Errorf("output '%s' is not a valid number", output)
	}

	return floatVal, nil
}
