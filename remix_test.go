package aurora

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gernest/render"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var pass = "mamamia"

func TestRemix_Home(t *testing.T) {
	ts, client, _ := testServer(t)
	defer ts.Close()

	res, err := client.Get(ts.URL)
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res.StatusCode)
	}
	w := &bytes.Buffer{}
	io.Copy(w, res.Body)
	if !contains(w.String(), "prove it yourself") {
		t.Error("Expected InSession not to be pset")
	}
}

func TestRemix_Register(t *testing.T) {
	ts, client, rx := testServer(t)
	defer ts.Close()
	registerURL := fmt.Sprintf("%s/auth/register", ts.URL)

	// get the form
	res1, err := client.Get(registerURL)
	if err != nil {
		t.Error(err)
	}
	if res1.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res1.StatusCode)
	}

	// Failing validation
	usr2 := url.Values{
		"first_name":    {"gernest"},
		"lastname":      {"aurora"},
		"email_address": {"gernest@aurora.com"},
		"pass":          {"ringadongdilo"},
		"confirm_pass":  {"ringadondilo"},
	}
	res2, err := client.PostForm(registerURL, usr2)
	defer res2.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res2.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res2.StatusCode)
	}

	// a valid form
	usr := url.Values{
		"first_name":    {"gernest"},
		"last_name":     {"aurora"},
		"email_address": {"gernest@aurora.com"},
		"pass":          {pass},
		"confirm_pass":  {pass},
	}

	// case there is an issue with db
	rx.cfg.AccountsBucket = ""
	res3, err := client.PostForm(registerURL, usr)
	defer res3.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res3.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected %d got %d", http.StatusFound, res3.StatusCode)
	}
	rx.cfg.AccountsBucket = "accounts" //Restore our config

	// case everything is ok
	res5, err := client.PostForm(registerURL, usr)
	defer res5.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res5.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res5.StatusCode)
	}
	w := &bytes.Buffer{}
	io.Copy(w, res5.Body)
	if !contains(w.String(), "search") {
		t.Error("Expected InSession to be set")
	}

	// case session is not new it should redirects to login path
	res6, err := client.PostForm(registerURL, usr)
	defer res5.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res6.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res6.StatusCode)
	}
	w.Reset() // reuse this buffer
	io.Copy(w, res6.Body)
	if !contains(w.String(), "login-form") {
		t.Errorf("Expected login form")
	}

	// making sure our password was encrypted
	user, err := GetUser(setDB(rx.db, rx.cfg.AccountsDB), rx.cfg.AccountsBucket, "gernest@aurora.com")
	if err != nil {
		t.Error(err)
	}
	err = verifyPass(user.Pass, pass)
	if err != nil {
		t.Error(err)
	}
}

func TestRemix_Login(t *testing.T) {
	ts, client, _ := testServer(t)
	defer ts.Close()
	loginURL := fmt.Sprintf("%s/auth/login", ts.URL)

	// get the login form
	re, err := client.Get(loginURL)
	if err != nil {
		t.Error(err)
	}
	if re.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, re.StatusCode)
	}

	// invalid form
	lform := url.Values{
		"email":    {"bogus"},
		"password": {"myass"},
	}
	res1, err := client.PostForm(loginURL, lform)
	defer res1.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res1.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res1.StatusCode)
	}
	w := &bytes.Buffer{}
	io.Copy(w, res1.Body)
	if !contains(w.String(), "login-form") {
		t.Error("Expected login form")
	}

	// case no such user but valid form
	lform.Set("email", "gernesti@aurora.com")
	lform.Set("password", "heydollringadongdillo")
	res2, err := client.PostForm(loginURL, lform)
	defer res1.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res2.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res2.StatusCode)
	}
	w.Reset()
	io.Copy(w, res2.Body)

	if !contains(w.String(), "login-form") {
		t.Error("Expected login form")
	}

	// wrong password
	lform.Set("email", "gernest@aurora.com")
	lform.Set("password", "heydollringadongdilloo")
	res3, err := client.PostForm(loginURL, lform)
	defer res3.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res2.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res3.StatusCode)
	}
	w.Reset()
	io.Copy(w, res3.Body)
	if !contains(w.String(), "login-form") {
		t.Error("Expected login form")
		t.Error(w.String())
	}

	// case everything is ok, it should redirect to the path specified in Remix.cfg
	rEmail := "gernest@aurora.com"
	lform.Set("email", rEmail)
	lform.Set("password", pass)
	res4, err := client.PostForm(loginURL, lform)
	defer res4.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res4.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res4.StatusCode)
	}
	w.Reset()
	io.Copy(w, res4.Body)

	if !contains(w.String(), "search") {
		t.Error("Expected InSession to be set")
	}
}

func TestRemix_Uploads(t *testing.T) {
	ts, client, rx := testServer(t)
	defer ts.Close()
	lform := url.Values{
		"email":    {"gernest@aurora.com"},
		"password": {pass},
	}
	loginURL := fmt.Sprintf("%s/auth/login", ts.URL)
	upURL := fmt.Sprintf("%s/uploads", ts.URL)
	res, err := client.PostForm(loginURL, lform)
	if err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Ecpected %d got %d", http.StatusOK, res.StatusCode)
	}
	defer res.Body.Close()
	buf, cType := testUpData("me.jpg", "single", t)
	res2, err := client.Post(upURL, cType, buf)
	if err != nil {
		t.Error(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Errorf("Ecpected %d got %d", http.StatusOK, res2.StatusCode)
	}
	w := &bytes.Buffer{}
	io.Copy(w, res2.Body)
	if !contains(w.String(), "jpg") {
		t.Errorf("Expected to save jpg file got %s", w.String())
	}

	buf, cType = testUpData("me.jpg", "multi", t)
	res3, err := client.Post(upURL, cType, buf)
	if err != nil {
		t.Error(err)
	}
	defer res3.Body.Close()
	if res3.StatusCode != http.StatusOK {
		t.Errorf("Ecpected %d got %d", http.StatusOK, res3.StatusCode)
	}
	w.Reset()
	io.Copy(w, res3.Body)
	if !contains(w.String(), "jpg") {
		t.Errorf("Expected to save jpg file got %s", w.String())
	}

	bAb := rx.cfg.AccountsBucket
	rx.cfg.AccountsBucket = ""
	buf, cType = testUpData("me.jpg", "single", t)
	res4, err := client.Post(upURL, cType, buf)
	if err != nil {
		t.Error(err)
	}
	defer res4.Body.Close()
	if res4.StatusCode != http.StatusInternalServerError {
		t.Errorf("Ecpected %d got %d", http.StatusInternalServerError, res4.StatusCode)
	}
	w.Reset()
	io.Copy(w, res4.Body)
	if !contains(w.String(), "bucket") {
		t.Errorf("Expected to save jpg file got %s", w.String())
	}
	rx.cfg.AccountsBucket = bAb

	bAb = rx.cfg.ProfilePicField
	rx.cfg.ProfilePicField = ""
	buf, cType = testUpData("me.jpg", "single", t)
	res5, err := client.Post(upURL, cType, buf)
	if err != nil {
		t.Error(err)
	}
	defer res5.Body.Close()
	if res4.StatusCode != http.StatusInternalServerError {
		t.Errorf("Ecpected %d got %d", http.StatusInternalServerError, res5.StatusCode)
	}
	w.Reset()
	io.Copy(w, res5.Body)
	if !contains(w.String(), " no such file") {
		t.Errorf("Expected  %s to contain no such file", w.String())
	}
	rx.cfg.ProfilePicField = bAb

}

func TestRemixt_ServeImages(t *testing.T) {
	ts, client, rx := testServer(t)
	defer ts.Close()

	email := "gernest@aurora.com"
	user, err := GetUser(setDB(rx.db, rx.cfg.AccountsDB), rx.cfg.AccountsBucket, email)
	if err != nil {
		t.Error(err)
	}
	pdb := getProfileDatabase(rx.cfg.DBDir, user.UUID, rx.cfg.DBExtension)
	p, err := GetProfile(setDB(rx.db, pdb), rx.cfg.ProfilesBucket, user.UUID)
	if err != nil {
		t.Error(err)
	}
	if len(p.Photos) != 3 {
		t.Errorf("Expected 3 got %d", len(p.Photos))
	}
	imgURL := fmt.Sprintf("%s/imgs?%s", ts.URL, p.Picture.Query)

	res, err := client.Get(imgURL)
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res.StatusCode)
	}
}

// This cleans up all the remix based test databases
func TestClean_remix(t *testing.T) {
	clenUp(t)
}

// Creates a test druve server for using the Remix handlers., it also returns a ready
// to use client, that supports sessions.
func testServer(t *testing.T) (*httptest.Server, *http.Client, *Remix) {
	cfg := &RemixConfig{
		AccountsBucket:  "accounts",
		SessionName:     "aurora",
		LoginRedirect:   "/",
		DBDir:           "fixture",
		DBExtension:     ".bdb",
		AccountsDB:      "fixture/accounts.bdb",
		ProfilesBucket:  "profiles",
		SessionsDB:      "fixture/sessions.bdb",
		SessionsBucket:  sBucket,
		ProfilePicField: "profile",
		PhotosField:     "photos",
	}

	rOpts := render.Options{
		Directory:     "templates",
		Extensions:    []string{".tmpl", ".html", ".tpl"},
		IsDevelopment: true,
	}
	sOpts := &sessions.Options{MaxAge: maxAge, Path: sPath}
	store := NewSessStore(setDB(testDb, cfg.SessionsDB), cfg.SessionsBucket, 10, sOpts, secret)
	rx := &Remix{
		db:    testDb,
		sess:  store,
		rendr: render.New(rOpts),
		cfg:   cfg,
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{Jar: jar}
	h := mux.NewRouter()
	h.HandleFunc("/", rx.Home)
	h.HandleFunc("/auth/register", rx.Register)
	h.HandleFunc("/auth/login", rx.Login).Methods("GET", "POST")
	h.HandleFunc("/imgs", rx.ServeImages).Methods("GET")
	h.HandleFunc("/uploads", rx.Uploads)
	ts := httptest.NewServer(h)
	return ts, client, rx
}

// checkts if the given str contains substring subStr
func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// deletes test database files
func clenUp(t *testing.T) {
	ts, _, rx := testServer(t)
	defer ts.Close()
	ferr := filepath.Walk(rx.cfg.DBDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == rx.cfg.DBExtension {
			err = os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if ferr != nil {
		t.Error(ferr)
	}
}
