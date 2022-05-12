// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package utils

import "fmt"

// MakeNiceJSON exists to address https://github.com/go-yaml/yaml/issues/139
func MakeNiceJSON(in map[string]interface{}) map[string]interface{} {
	tmp := in
	for k, v := range tmp {
		tmp[k] = convertKeys(v)
	}

	return tmp
}

// recursive function to deal with all the types.
func convertKeys(in interface{}) interface{} {
	switch in := in.(type) {
	case []interface{}:
		res := make([]interface{}, len(in))
		for i, v := range in {
			res[i] = convertKeys(v)
		}

		return res
	case map[interface{}]interface{}:
		res := make(map[string]interface{})
		for k, v := range in {
			res[fmt.Sprintf("%v", k)] = convertKeys(v)
		}

		return res
	default:
		return in
	}
}
