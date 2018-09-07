package version

import "testing"

func TestString(t *testing.T) {
	tests := []struct {
		name    string
		version string
		hash    string
		want    string
	}{
		{"Dev version - no version info supplied", "", "", "Version : Dev build (no valid version)\nGit Hash: N/A\n"},
		{"Proper version", "v0.1.0", "c0ff33", "Version : v0.1.0\nGit Hash: c0ff33\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			GitHash = tt.hash
			if got := String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
