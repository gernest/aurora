package aurora

import (
	"testing"

	"github.com/gernest/nutz"
)

var testDb = nutz.NewStorage("test.ddb", 0600, nil)
var db = nutz.NewStorage("auth_test.ddb", 0600, nil)

func TestCreateAccount(t *testing.T) {
	bucket := "test_create_account"
	dataset := []struct {
		uuid, email string
	}{
		{"db0668ac-7eba-40dd-56ee-0b1c0b9b415d", "gernest@aurora.com"},
		{"e6917dfe-b4f6-49b8-5628-83dd2a430e9a", "gernest@aurora.tz"},
		{"bc5288cf-4120-4f3c-5957-b19e093a12f4", "gernest@aurora.io"},
	}
	for _, u := range dataset {
		usr := &User{
			EmailAddress: u.email,
			UUID:         u.uuid,
		}
		if err := CreateAccount(db, usr, bucket); err != nil {
			t.Error(err)
		}
	}
}

func TestGetUser(t *testing.T) {
	bucket := "test_get"
	dataset := []struct {
		uuid, email string
	}{
		{"db0668ac-7eba-40dd-56ee-0b1c0b9b415d", "gernest@aurora.com"},
		{"e6917dfe-b4f6-49b8-5628-83dd2a430e9a", "gernest@aurora.tz"},
		{"bc5288cf-4120-4f3c-5957-b19e093a12f4", "gernest@aurora.io"},
	}
	for _, u := range dataset {
		usr := &User{
			EmailAddress: u.email,
			UUID:         u.uuid,
		}
		if err := CreateAccount(db, usr, bucket); err != nil {
			t.Error(err)
		}
	}
	for _, u := range dataset {
		user, err := GetUser(db, bucket, u.email)
		if err != nil {
			t.Errorf("geeting user %v", err)
		}
		if user.UUID != u.uuid {
			t.Errorf("expected %s got %s", u.uuid, user.UUID)
		}
		if user.EmailAddress != u.email {
			t.Errorf("expected %s got %s", u.email, user.EmailAddress)
		}
	}
}
func TestGetAll(t *testing.T) {
	var origin string
	bucket := "get_all"
	defer db.DeleteDatabase()
	for _ = range []int{1, 2, 3, 4} {
		usr := &User{
			UUID: getUUID(),
		}
		usr.EmailAddress = usr.UUID
		err := CreateAccount(db, usr, bucket)
		if err != nil {
			t.Error(err)
		}
		origin = origin + "," + usr.UUID
	}
	curr, err := GetAllUsers(db, bucket)
	if err != nil {
		t.Error(err)
	}
	for _, v := range curr {
		if !contains(origin, v) {
			t.Errorf("expected %s to be in %s", v, origin)
		}
	}
	zz, err := GetAllUsers(testDb, "lora")
	if err == nil {
		t.Error("expected an error")
	}
	if zz != nil {
		t.Errorf("expected nil got %v", zz)
	}
}
