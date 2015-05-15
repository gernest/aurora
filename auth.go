package aurora

import (
	"encoding/json"

	"github.com/gernest/nutz"
)

// CreateAccount creates a new account, where id will be the value returned by
// invoking Email() method.
func CreateAccount(db nutz.Storage, a Account, bucket string) error {
	return createIfNotexist(db, a, bucket, a.Email())
}

// GetUser retrives a user.
func GetUser(db nutz.Storage, bucket, email string, nest ...string) (*User, error) {
	usr := &User{}
	err := getAndUnmarshall(db, bucket, email, usr)
	if err != nil {
		return nil, err
	}
	return usr, nil
}

// GetAllUsers returns a slice of all users.
func GetAllUsers(db nutz.Storage, bucket string, nest ...string) ([]string, error) {
	var usrs []string
	d := db.GetAll(bucket, nest...)
	if d.Error != nil {
		return nil, d.Error
	}
	for _, v := range d.DataList {
		us := &User{}
		err := json.Unmarshal(v, us)
		if err != nil {
			// log thus
		}
		if err == nil {
			usrs = append(usrs, us.UUID)
		}
	}
	return usrs, nil
}
