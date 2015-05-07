package aurora

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var (
	maxAge  = 30
	sPath   = "/"
	cName   = "youngWarlock"
	sBucket = "sessions"
	secret  = []byte("my-secret")
	testURL = "http://www.example.com"
)

func TestSession_New(t *testing.T) {
	store, req := sessSetup(t)
	testNewSess(store, req, t)
}

func TestSession_Save(t *testing.T) {
	store, req := sessSetup(t)
	testNewSess(store, req, t)
	testSaveSess(store, req, t, "user", "gernest")
}

func TestSess_Get(t *testing.T) {
	opts := &sessions.Options{MaxAge: maxAge, Path: sPath}
	store, req := sessSetup(t)
	s := testSaveSess(store, req, t, "user", "gernest")
	c, err := securecookie.EncodeMulti(s.Name(), s.ID, securecookie.CodecsFromPairs(secret)...)
	if err != nil {
		t.Error(err)
	}
	newCookie := sessions.NewCookie(s.Name(), c, opts)
	req.AddCookie(newCookie)
	s, err = store.New(req, cName)
	if err != nil {
		t.Error(err)
	}
	if s.IsNew {
		t.Errorf("Expected  false, actual %v", s.IsNew)
	}
	ss, err := store.Get(req, cName)
	if err != nil {
		t.Error(err)
	}
	if ss.IsNew {
		t.Errorf("Expected  false, actual %v", ss.IsNew)
	}
	if ss.Values["user"] != "gernest" {
		t.Errorf("Expected gernest, actual %s", ss.Values["user"])
	}
}
func TestSess_Delete(t *testing.T) {
	store, req := sessSetup(t)
	s := testSaveSess(store, req, t, "user", "gernest")
	defer testDb.DeleteDatabase()
	w := httptest.NewRecorder()
	err := store.Delete(req, w, s)
	if err != nil {
		t.Error(err)
	}
}

func sessSetup(t *testing.T) (*Session, *http.Request) {
	opts := &sessions.Options{MaxAge: maxAge, Path: sPath}
	store := NewSessStore(testDb, sBucket, 10, opts, secret)
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		t.Error(err)
	}
	return store, req
}

func testNewSess(ss *Session, req *http.Request, t *testing.T) *sessions.Session {
	s, err := ss.New(req, cName)
	if err == nil {
		if !s.IsNew {
			t.Errorf("Expected true actual %v", s.IsNew)
		}
		t.Errorf("Expected \"http: named cookie not present\" actual nil")
	}
	return s
}
func testSaveSess(ss *Session, req *http.Request, t *testing.T, key, val string) *sessions.Session {
	s := testNewSess(ss, req, t)
	s.Values[key] = val
	w := httptest.NewRecorder()
	err := s.Save(req, w)
	if err != nil {
		t.Error(err)
	}
	return s
}
