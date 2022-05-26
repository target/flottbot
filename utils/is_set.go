// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

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
