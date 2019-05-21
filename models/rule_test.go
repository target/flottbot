package models

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestRuleValidation(t *testing.T) {

	os.Setenv("FB_ENV", "dev")
	defer os.Unsetenv("FB_ENV")

	r := new(Rule)
	r.Name = "test"
	r.OutputToRooms = []string{"operations-${FB_ENV}"}

	bot := new(Bot)
	bot.Log = *logrus.StandardLogger()
	bot.Log.SetLevel(logrus.DebugLevel)

	err := r.Validate(bot)
	if err != nil {
		t.Error(err)
	}

	if r.OutputToRooms[0] != "operations-dev" {
		t.Errorf("expected %s but go %s", "operations-dev", r.OutputToRooms[0])
	}
}
