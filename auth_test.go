package aurora

import (
	"testing"

	"github.com/gernest/nutz"
)

var testDb = nutz.NewStorage("fixture/test.ddb", 0600, nil)

func TestCreateAccount(t *testing.T) {
	var aBucket = "accounts"
	usr := NewUser()
	usr.EmailAddress = "gernest@mwanza.tz"
	if err := CreateAccount(testDb, usr, aBucket); err != nil {
		t.Error(err)
	}
}

func TestGetUser(t *testing.T) {
	var aBucket = "accounts"
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
func TestGetAll(t *testing.T) {
	var (
		testBucket = "test buck"
		origin     string
		curr       []string
		err        error
	)

	for _ = range []int{1, 2, 3, 4} {
		usr := &User{
			UUID: getUUID(),
		}
		usr.EmailAddress = usr.UUID
		err = CreateAccount(testDb, usr, testBucket)
		if err != nil {
			t.Error(err)
		}
		origin = origin + "," + usr.UUID
	}
	curr, err = GetAllUsers(testDb, testBucket)
	if err != nil {
		t.Error(err)
	}
	for _, v := range curr {
		if !contains(origin, v) {
			t.Errorf("Expected %s to be in %s", v, origin)
		}
	}
	zz, err := GetAllUsers(testDb, "lora")
	if err == nil {
		t.Error("Expected an error")
	}
	if zz != nil {
		t.Errorf("Expected nil got %v", zz)
	}
}

// remove the database used by the above tests
func TestClean_auth(t *testing.T) {
	testDb.DeleteDatabase()
}
