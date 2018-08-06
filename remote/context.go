package remote

import "context"

const key = "remote"

// FromContext returns the Remote associated with this context
func FromContext(c context.Context) Remote {
	return c.Value(key).(Remote)
}
