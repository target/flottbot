// SPDX-License-Identifier: Apache-2.0

package utils

import "strings"

// IsSet is a helper function to check whether any of the supplied
// strings are empty or unsubstituted (ie. still in ${<string>} format).
func IsSet(s ...string) bool {
	for _, v := range s {
		if v == "" || strings.HasPrefix(v, "${") {
			return false
		}
	}

	return true
}
