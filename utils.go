package aurora

import (
	"encoding/json"
	"errors"
	"log"
	"path/filepath"

	"github.com/nu7hatch/gouuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/gernest/nutz"
	"github.com/gorilla/sessions"
)

// serialize the given object obj into json format and saves it into the dtabase
func marshalAndCreate(db nutz.Storage, obj interface{}, buck, key string, nest ...string) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	c := db.Create(buck, key, data, nest...)
	return c.Error
}

// serialize the given object to json and saves it into the database
func marshalAndUpdate(db nutz.Storage, obj interface{}, buck, key string, nest ...string) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	c := db.Update(buck, key, data, nest...)
	return c.Error
}

// serialize and saves th object to the database, but checks first if the key already exist.
// When there is already a record with a given key an error is returned.
func createIfNotexist(db nutz.Storage, obj interface{}, buck, key string, nest ...string) error {
	if g := db.Get(buck, key, nest...); g.Error != nil {
		return marshalAndCreate(db, obj, buck, key, nest...)
	}
	return errors.New("aurora: already exist")
}

// Encrypts a given string using bcrypt library. It returns the hashed password as a string,
// or any error
func hashPassword(pass string) (string, error) {
	np, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(np), err
}

// Flash is a helper for storing and retrieving flash messages
// TODO : Move this to another file, it just don't look like it belongs here
type Flash struct {
	Data map[string]interface{}
}

// NewFlash creates a new flash
func NewFlash() *Flash {
	return &Flash{Data: make(map[string]interface{})}
}

// Success adds a success message
func (f *Flash) Success(msg string) {
	f.Data["FlashSuccess"] = msg
}

// Notice adds a notice flash message
func (f *Flash) Notice(msg string) {
	f.Data["FlashNotice"] = msg
}

// Error adds an error message
func (f *Flash) Error(msg string) {
	f.Data["FlashError"] = msg
}

// Add saves the flsah to the given session
func (f *Flash) Save(s *sessions.Session) {
	data, err := json.Marshal(f)
	if err == nil {
		s.AddFlash(data)
	}
}

// Get retrieves any flash messages in the session
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

// Retrives data from the dataase, and marshalls the result to the given obj. Thhis
// uses json decoding.
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

// Verfiess if the given has mathes the password. The hash must be a bcrypt encoded has.
// it uses bcrypt to compare the two passwords
func verifyPass(hash, pass string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
}

// returns a new UUIDv4 string
func getUUID() string {
	id, err := uuid.NewV4()
	if err != nil {
		// TODO :log
	}
	return id.String()
}

func getProfileDatabase(dbDir, profileID, dbExt string) string {
	return filepath.Join(dbDir, profileID+dbExt)
}
