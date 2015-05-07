package aurora

import (
	"testing"

	"github.com/gernest/nutz"
)

var (
	testDb  = nutz.NewStorage("test.db", 0600, nil)
	aBucket = "accounts"
)

func TestCreateAccount(t *testing.T) {
	usr := NewUser()
	usr.EmailAdress = "gernest@mwanza.tz"
	if err := CreateAccount(testDb, usr, aBucket); err != nil {
		t.Error(err)
	}
}

func TestCleanUp(t *testing.T) {
	testDb.DeleteDatabase()
}
