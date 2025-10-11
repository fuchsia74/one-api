package common

import "testing"

func TestExpandLogDirPath(t *testing.T) {
	t.Setenv("APP_ROOT", "/srv/app")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unix style",
			input:    "$APP_ROOT/logs",
			expected: "/srv/app/logs",
		},
		{
			name:     "windows style with known default",
			input:    "%DATA_DIR%/logs",
			expected: "/data/logs",
		},
		{
			name:     "unknown windows style passthrough",
			input:    "%UNKNOWN_VAR%/logs",
			expected: "%UNKNOWN_VAR%/logs",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := expandLogDirPath(tc.input)
			if got != tc.expected {
				t.Fatalf("expandLogDirPath(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
