package aurora

import (
	"encoding/json"
	"errors"

	"github.com/gernest/nutz"
)

func marshalAndCreate(db nutz.Storage, obj interface{}, buck, key string, nest ...string) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	c := db.Create(buck, key, data, nest...)
	return c.Error
}

func marshalAndUpdate(db nutz.Storage, obj interface{}, buck, key string, nest ...string) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	c := db.Update(buck, key, data, nest...)
	return c.Error
}

func createIfNotexist(db nutz.Storage, obj interface{}, buck, key string, nest ...string) error {
	if g := db.Get(buck, key, nest...); g.Error != nil {
		return marshalAndCreate(db, obj, buck, key, nest...)
	}
	return errors.New("aurora: already exist")
}
