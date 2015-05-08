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
	usr.EmailAddress = "gernest@mwanza.tz"
	if err := CreateAccount(testDb, usr, aBucket); err != nil {
		t.Error(err)
	}
}

func TestGetUser(t *testing.T) {
	usr := NewUser()
	usr.EmailAddress = "geo@mwanza.tz"
	if err := CreateAccount(testDb, usr, aBucket); err != nil {
		t.Error(err)
	}
	user, err := GetUser(testDb, aBucket, usr.EmailAddress)
	if err != nil {
		t.Error(err)
	}
	if err == nil {
		if user.EmailAddress != usr.EmailAddress {
			t.Errorf("Expected %s got %s", usr.EmailAddress, user.EmailAddress)
		}
	}

}
func TestCleanUp(t *testing.T) {
	testDb.DeleteDatabase()
}
