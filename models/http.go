// SPDX-License-Identifier: Apache-2.0

package models

// HTTPResponse base HTTP response data structure.
type HTTPResponse struct {
	Status int
	Raw    string
	Data   any
}
