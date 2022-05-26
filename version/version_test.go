// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package version

import "testing"

func TestString(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{"Dev version - no version info supplied", "", "Version : dev\nGit Hash: unknown\n"},
		{"Proper version", "v0.1.0", "Version : v0.1.0\nGit Hash: unknown\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			if got := String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
