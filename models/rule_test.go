package models

import (
	"os"
	"testing"
)

func TestRuleValidation(t *testing.T) {

	os.Setenv("FB_ENV", "dev")
	defer os.Unsetenv("FB_ENV")

	r := new(Rule)
	r.OutputToRooms = []string{"operations-${FB_ENV}"}

	err := r.Validate()
	if err != nil {
		t.Error(err)
	}

	if r.OutputToRooms[0] != "operations-dev" {
		t.Errorf("expected %s but go %s", "operations-dev", r.OutputToRooms[0])
	}
}
