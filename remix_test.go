package aurora

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
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
	if err == nil {
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, res.StatusCode)
		}
		w := &bytes.Buffer{}
		io.Copy(w, res.Body)
		if !contains(w.String(), "prove it yourself") {
			t.Error("Expected InSession not to be pset")
		}
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
	if err == nil {
		if res1.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, res1.StatusCode)
		}
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
	if err == nil {
		if res3.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected %d got %d", http.StatusFound, res3.StatusCode)
		}
	}

	rx.cfg.AccountsBucket = "accounts" //Restore our config

	//case our session is messe up
	//backUp := *rx.sess
	//sOpts := &sessions.Options{MaxAge: maxAge, Path: sPath}
	//ns := NewSessStore(testDb, "", 10, sOpts, secret)
	//rx.sess = ns
	//res4, err := client.PostForm(registerURL, usr)
	//defer res4.Body.Close()
	//if err != nil {
	//	t.Error(err)
	//}
	//if err == nil {
	//	if res4.StatusCode != http.StatusInternalServerError {
	//		t.Errorf("Expected %d got %d", http.StatusInternalServerError, res4.StatusCode)
	//	}
	//}
	//rx.sess = &backUp

	// case everything is ok
	res5, err := client.PostForm(registerURL, usr)
	defer res5.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if err == nil {
		if res5.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, res5.StatusCode)
		}
		w := &bytes.Buffer{}
		io.Copy(w, res5.Body)
		if !contains(w.String(), "search") {
			t.Error("Expected InSession to be set")
		}
	}

	// case session is not new it should redirects to login path
	res6, err := client.PostForm(registerURL, usr)
	defer res5.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if err == nil {
		if res6.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, res6.StatusCode)
		}
		w := &bytes.Buffer{}
		io.Copy(w, res6.Body)
		if !contains(w.String(), "login-form") {
			t.Errorf("Expected login form")
		}
	}

	// making sure our password was encrypted
	user, err := GetUser(testDb, rx.cfg.AccountsBucket, "gernest@aurora.com")
	if err != nil {
		t.Error(err)
	}
	if err == nil {
		err = verifyPass(user.Pass, pass)
		if err != nil {
			t.Error(err)
		}
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
	if err == nil {
		if re.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, re.StatusCode)
		}
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
	if err == nil {
		if res1.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, res1.StatusCode)
		}
		w := &bytes.Buffer{}
		io.Copy(w, res1.Body)

		if !contains(w.String(), "login-form") {
			t.Error("Expected login form")
		}
	}

	// case no but valid form
	lform.Set("email", "gernesti@aurora.com")
	lform.Set("password", "heydollringadongdillo")
	res2, err := client.PostForm(loginURL, lform)
	defer res1.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if err == nil {
		if res2.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, res2.StatusCode)
		}
		w := &bytes.Buffer{}
		io.Copy(w, res2.Body)

		if !contains(w.String(), "login-form") {
			t.Error("Expected login form")
		}
	}

	// wrong password
	lform.Set("email", "gernest@aurora.com")
	lform.Set("password", "heydollringadongdilloo")
	res3, err := client.PostForm(loginURL, lform)
	defer res3.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if err == nil {
		if res2.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, res3.StatusCode)
		}
		w := &bytes.Buffer{}
		io.Copy(w, res3.Body)

		if !contains(w.String(), "login-form") {
			t.Error("Expected login form")
			t.Error(w.String())
		}
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
	if err == nil {
		if res4.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, res4.StatusCode)
		}
		w := &bytes.Buffer{}
		io.Copy(w, res4.Body)

		if !contains(w.String(), "search") {
			t.Error("Expected InSession to be set")
		}
	}
}

func testServer(t *testing.T) (*httptest.Server, *http.Client, *Remix) {
	cfg := &RemixConfig{
		AccountsBucket: "accounts",
		SessionName:    "aurora",
		LoginRedirect:  "/",
	}
	rOpts := render.Options{
		Directory:     "templates",
		Extensions:    []string{".tmpl", ".html", ".tpl"},
		IsDevelopment: true,
	}
	sOpts := &sessions.Options{MaxAge: maxAge, Path: sPath}
	store := NewSessStore(testDb, sBucket, 10, sOpts, secret)
	rx := &Remix{
		sess:       store,
		rendr:      render.New(rOpts),
		cfg:        cfg,
		accoundtDB: testDb,
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
	ts := httptest.NewServer(h)
	return ts, client, rx
}

func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}
