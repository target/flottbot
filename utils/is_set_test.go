// SPDX-License-Identifier: Apache-2.0

package utils

import "testing"

func TestIsSet(t *testing.T) {
	type args struct {
		s []string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"empty",
			args{[]string{""}},
			false,
		},
		{
			"unset/unsubstituted",
			args{[]string{"${foo}"}},
			false,
		},
		{
			"set",
			args{[]string{"bar"}},
			true,
		},
		{
			"many - one unset",
			args{[]string{"foo", "bar", "${foo}"}},
			false,
		},
		{
			"many - all set",
			args{[]string{"foo", "bar"}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSet(tt.args.s...); got != tt.want {
				t.Errorf("IsSet() = %v, want %v", got, tt.want)
			}
		})
	}
}
