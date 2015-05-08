package aurora

import "github.com/gernest/nutz"

func CreateAccount(db nutz.Storage, a Account, bucket string) error {
	return createIfNotexist(db, a, bucket, a.Email())
}

func GetUser(db nutz.Storage, bucket, email string, nest ...string) (*User, error) {
	usr := &User{}
	err := getAndUnmarshall(db, bucket, email, usr)
	if err != nil {
		return nil, err
	}
	return usr, nil
}
