// Copyright (c) 2023 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package utils

import "fmt"

// MakeNiceJSON exists to address https://github.com/go-yaml/yaml/issues/139
func MakeNiceJSON(in map[string]any) map[string]any {
	tmp := in
	for k, v := range tmp {
		tmp[k] = convertKeys(v)
	}

	return tmp
}

// recursive function to deal with all the types.
func convertKeys(in any) any {
	switch in := in.(type) {
	case []any:
		res := make([]any, len(in))
		for i, v := range in {
			res[i] = convertKeys(v)
		}

		return res
	case map[any]any:
		res := make(map[string]any)
		for k, v := range in {
			res[fmt.Sprintf("%v", k)] = convertKeys(v)
		}

		return res
	default:
		return in
	}
}
