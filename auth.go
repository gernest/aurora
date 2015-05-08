package aurora

import "github.com/gernest/nutz"

// CreateAccount creates a new record in the bucket, using email as key
func CreateAccount(db nutz.Storage, a Account, bucket string) error {
	return createIfNotexist(db, a, bucket, a.Email())
}

// GetUser retrives a user from the database
func GetUser(db nutz.Storage, bucket, email string, nest ...string) (*User, error) {
	usr := &User{}
	err := getAndUnmarshall(db, bucket, email, usr)
	if err != nil {
		return nil, err
	}
	return usr, nil
}
