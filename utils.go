package aurora

import (
	"encoding/json"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"

	"github.com/gernest/nutz"
	"github.com/gorilla/sessions"
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

func hashPassword(pass string) (string, error) {
	np, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(np), err
}

type Flash struct {
	Data map[string]interface{}
}

func NewFlash() *Flash {
	return &Flash{Data: make(map[string]interface{})}
}
func (f *Flash) Success(msg string) {
	f.Data["FlashSuccess"] = msg
}

func (f *Flash) Notice(msg string) {
	f.Data["FlashNotice"] = msg
}

func (f *Flash) Error(msg string) {
	f.Data["FlashError"] = msg
}
func (f *Flash) Add(s *sessions.Session) {
	data, err := json.Marshal(f)
	if err == nil {
		s.AddFlash(data)
	}
}

func (f *Flash) Get(s *sessions.Session) *Flash {
	if flashes := s.Flashes(); flashes != nil {
		data := flashes[0]
		if err := json.Unmarshal(data.([]byte), f); err != nil {
			log.Println(err)
			return nil
		}
		return f
	}
	return nil
}

func getAndUnmarshall(db nutz.Storage, bucket, key string, obj interface{}, nest ...string) error {
	g := db.Get(bucket, key, nest...)
	if g.Error != nil {
		return g.Error
	}
	err := json.Unmarshal(g.Data, obj)
	if err != nil {
		return err
	}
	return nil
}

func verifyPass(hash, pass string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
}
