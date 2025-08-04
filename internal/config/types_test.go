package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tiernacity/ratchet/internal/errors"
)

func TestComparisonOperator_String(t *testing.T) {
	tests := []struct {
		name string
		op   ComparisonOperator
		want string
	}{
		{"greater than", OpGreaterThan, ">"},
		{"greater than or equal", OpGreaterThanOrEqual, ">="},
		{"equal", OpEqual, "=="},
		{"less than or equal", OpLessThanOrEqual, "<="},
		{"less than", OpLessThan, "<"},
		{"unknown", OpUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.op.String())
		})
	}
}

func TestComparisonOperator_HumanString(t *testing.T) {
	tests := []struct {
		name string
		op   ComparisonOperator
		want string
	}{
		{"greater than", OpGreaterThan, "greater than"},
		{"greater than or equal", OpGreaterThanOrEqual, "greater than or equal to"},
		{"equal", OpEqual, "equal to"},
		{"less than or equal", OpLessThanOrEqual, "less than or equal to"},
		{"less than", OpLessThan, "less than"},
		{"unknown", OpUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.op.HumanString())
		})
	}
}

func TestConfig_Normalize(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		want    ComparisonOperator
		wantErr bool
		errMsg  string
	}{
		{
			name: "single operator - greater than",
			config: Config{
				GreaterThan: "main",
			},
			want: OpGreaterThan,
		},
		{
			name: "single operator - less than",
			config: Config{
				LessThan: "main",
			},
			want: OpLessThan,
		},
		{
			name: "single operator - equal",
			config: Config{
				Equal: "main",
			},
			want: OpEqual,
		},
		{
			name: "multiple operators",
			config: Config{
				GreaterThan: "main",
				LessThan:    "main",
			},
			wantErr: true,
			errMsg:  "multiple comparison operators specified",
		},
		{
			name: "no operator - valid for no-comparison mode",
			config: Config{
				Metric: "echo 42",
			},
			want: OpUnknown,
		},
		{
			name: "all operators set",
			config: Config{
				GreaterThan:        "main",
				GreaterThanOrEqual: "main",
				Equal:              "main",
				LessThanOrEqual:    "main",
				LessThan:           "main",
			},
			wantErr: true,
			errMsg:  "multiple comparison operators specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Normalize()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, tt.config.Operator)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Metric:     "echo 42",
				BaseBranch: "main",
				Operator:   OpGreaterThan,
			},
		},
		{
			name: "missing metric command",
			config: Config{
				BaseBranch: "main",
				Operator:   OpGreaterThan,
			},
			wantErr: true,
			errMsg:  "metric command is required",
		},
		{
			name: "missing base branch with comparison operator",
			config: Config{
				Metric:   "echo 42",
				Operator: OpGreaterThan,
			},
			wantErr: true,
			errMsg:  "base branch cannot be empty when using comparison operator",
		},
		{
			name: "no-comparison mode valid",
			config: Config{
				Metric:   "echo 42",
				Operator: OpUnknown,
				// BaseBranch not required for no-comparison mode
			},
		},
		{
			name: "potentially dangerous command - rm",
			config: Config{
				Metric:     "rm -rf /tmp/test && echo 42",
				BaseBranch: "main",
				Operator:   OpGreaterThan,
			},
			// Should not error, just warn
		},
		{
			name: "potentially dangerous command - delete",
			config: Config{
				Metric:     "delete-old-files.sh | wc -l",
				BaseBranch: "main",
				Operator:   OpGreaterThan,
			},
			// Should not error, just warn
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := errors.NewValidationError("test-field", "test message")
	assert.Equal(t, "test message", err.Error())
}
