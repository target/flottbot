// Copyright (c) 2023 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package remote

import "context"

const key = "remote"

// FromContext returns the Remote associated with this context.
func FromContext(c context.Context) Remote {
	return c.Value(key).(Remote)
}
