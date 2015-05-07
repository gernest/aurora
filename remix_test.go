package aurora

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gernest/render"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func TestRemix_Register(t *testing.T) {
	ts, client, _ := testServer(t)
	defer ts.Close()
	registerURL := fmt.Sprintf("%s/auth/register", ts.URL)

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
	if res2.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected %d got %d", http.StatusInternalServerError, res2.StatusCode)
	}

	// Success register
	usr := url.Values{
		"first_name":    {"gernest"},
		"last_name":     {"aurora"},
		"email_address": {"gernest@aurora.com"},
		"pass":          {"ringadongdilo"},
		"confirm_pass":  {"ringadongdilo"},
	}
	res, err := client.PostForm(registerURL, usr)
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Expected %d got %d", http.StatusFound, res.StatusCode)
	}
}
func testServer(t *testing.T) (*httptest.Server, *http.Client, *Remix) {
	cfg := &RemixConfig{AccountsBucket: "accounts", SessionName: "aurora"}
	rOpts := render.Options{Directory: "fixture"}
	sOpts := &sessions.Options{MaxAge: maxAge, Path: sPath}
	store := NewSessStore(testDb, sBucket, 10, sOpts, secret)
	rendr := render.New(rOpts)
	rx := &Remix{
		sess:       store,
		rendr:      rendr,
		cfg:        cfg,
		accoundtDB: testDb,
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{Jar: jar}
	h := mux.NewRouter()
	h.HandleFunc("/auth/register", rx.Register)
	ts := httptest.NewServer(h)
	return ts, client, rx
}
