package aurora

import (
	"testing"

	"github.com/gernest/nutz"
)

var (
	testDb  = nutz.NewStorage("fixture/test.ddb", 0600, nil)
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
	if user.EmailAddress != usr.EmailAddress {
		t.Errorf("Expected %s got %s", usr.EmailAddress, user.EmailAddress)
	}
}

func TestClean_auth(t *testing.T) {
	testDb.DeleteDatabase()
}
