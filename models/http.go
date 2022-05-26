// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package models

// HTTPResponse base HTTP response data structure.
type HTTPResponse struct {
	Status int
	Raw    string
	Data   interface{}
}
