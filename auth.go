package aurora

import (
	"github.com/gernest/nutz"
)

func CreateAccount(db nutz.Storage, a Account, bucket string) error {
	return createIfNotexist(db, a, bucket, a.Email())
}
