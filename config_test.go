package gqfsl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckDefaultConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    Config
		expected Config
	}{
		{
			name: "Set all defaults",
			input: Config{
				Sql:  SqlConf{},
				Cron: CronConf{},
			},
			expected: Config{
				Sql: SqlConf{
					BaseName: "email.sqlite",
					Timeout:  5 * time.Second,
					Path:     ".",
				},
				Cron: CronConf{
					DurationRetry:           5 * time.Minute,
					DurationSaveFailMessage: 24 * time.Hour,
					CountTry:                100,
				},
			},
		},
		{
			name: "Partial defaults",
			input: Config{
				Sql: SqlConf{
					BaseName: "custom.sqlite",
					Timeout:  10 * time.Second,
				},
				Cron: CronConf{
					DurationRetry: 10 * time.Minute,
					CountTry:      50,
				},
			},
			expected: Config{
				Sql: SqlConf{
					BaseName: "custom.sqlite",
					Timeout:  10 * time.Second,
					Path:     ".",
				},
				Cron: CronConf{
					DurationRetry:           10 * time.Minute,
					DurationSaveFailMessage: 24 * time.Hour,
					CountTry:                50,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkDefaultConfig(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
