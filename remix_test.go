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
)

func TestRemix_Home(t *testing.T) {
	var (
		err error
		res *http.Response
	)
	ts, client, _ := testServer(t)
	defer ts.Close()

	res, err = client.Get(ts.URL)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res, http.StatusOK, "pitch")
	if err != nil {
		t.Error(err)
	}
}

func TestRemix_Register(t *testing.T) {
	var (
		registratinPath              = "/auth/register"
		pass                         = "mamamia"
		err                          error
		res1, res2, res3, res4, res5 *http.Response
		vars                         url.Values
	)

	ts, client, rx := testServer(t)
	defer ts.Close()

	registerURL := fmt.Sprintf("%s%s", ts.URL, registratinPath)

	// get the form
	res1, err = client.Get(registerURL)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res1, http.StatusOK)
	if err != nil {
		t.Error(err)
	}

	// Failing validation
	vars = url.Values{
		"first_name":    {"gernest"},
		"lastname":      {"aurora"},
		"email_address": {"gernest@aurora.com"},
		"pass":          {"ringadongdilo"},
		"confirm_pass":  {"ringadondilo"},
	}
	res2, err = client.PostForm(registerURL, vars)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res2, http.StatusOK)
	if err != nil {
		t.Error(err)
	}

	// a valid form
	vars = url.Values{
		"first_name":    {"gernest"},
		"last_name":     {"aurora"},
		"email_address": {"gernest@aurora.com"},
		"pass":          {pass},
		"confirm_pass":  {pass},
	}

	// case there is an issue with db
	rx.cfg.AccountsBucket = ""
	res3, err = client.PostForm(registerURL, vars)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res3, http.StatusInternalServerError)
	if err != nil {
		t.Error(err)
	}
	rx.cfg.AccountsBucket = "accounts" //Restore our config

	// case everything is ok
	res4, err = client.PostForm(registerURL, vars)
	if err != nil {
		t.Error(err)
	}

	err = checkResponse(res4, http.StatusOK, "search")
	if err != nil {
		t.Error(err)
	}

	// case session is not new it should redirects to login path
	res5, err = client.PostForm(registerURL, vars)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res5, http.StatusOK, "/auth/logout")
	if err != nil {
		t.Error(err)
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
	var (
		email                       = "gernest@aurora.com"
		loginPath                   = "/auth/login"
		pass                        = "mamamia"
		err                         error
		res, res1, res2, res3, res4 *http.Response
		vars                        url.Values
	)

	ts, client, _ := testServer(t)
	defer ts.Close()
	loginURL := fmt.Sprintf("%s%s", ts.URL, loginPath)

	// get the login form
	res, err = client.Get(loginURL)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res, http.StatusOK)
	if err != nil {
		t.Error(err)
	}

	// invalid form
	vars = url.Values{
		"email":    {"bogus"},
		"password": {"myass"},
	}
	res1, err = client.PostForm(loginURL, vars)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res1, http.StatusOK, "login-form")
	if err != nil {
		t.Error(err)
	}

	// case no such user but valid form
	vars = url.Values{
		"email":    {"gernesti@aurora.com"},
		"password": {"heydollringadongdillo"},
	}
	res2, err = client.PostForm(loginURL, vars)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res2, http.StatusOK, "login-form")
	if err != nil {
		t.Error(err)
	}

	// wrong password
	vars = url.Values{
		"email":    {"gernest@aurora.com"},
		"password": {"heydollringadongdilloo"},
	}
	res3, err = client.PostForm(loginURL, vars)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res3, http.StatusOK, "login-form")
	if err != nil {
		t.Error(err)
	}

	// case everything is ok, it should redirect to the path specified in Remix.cfg
	vars = url.Values{
		"email":    {email},
		"password": {pass},
	}
	res4, err = client.PostForm(loginURL, vars)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res4, http.StatusOK, "search")
	if err != nil {
		t.Error(err)
	}

}

func TestRemix_Uploads(t *testing.T) {
	var (
		w                                 = &bytes.Buffer{}
		uploadPath                        = "/uploads"
		loginPath                         = "/auth/login"
		pass                              = "mamamia"
		contentType                       string
		err                               error
		res, res0, res1, res2, res3, res4 *http.Response
		vars                              url.Values
		content                           *bytes.Buffer
	)
	ts, client, rx := testServer(t)
	defer ts.Close()
	vars = url.Values{
		"email":    {"gernest@aurora.com"},
		"password": {pass},
	}
	loginURL := fmt.Sprintf("%s%s", ts.URL, loginPath)
	upURL := fmt.Sprintf("%s%s", ts.URL, uploadPath)

	content, contentType = testUpData("me.jpg", "single", t)
	res0, err = client.Post(upURL, contentType, content)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res0, http.StatusForbidden, errForbidden.Error())
	if err != nil {
		t.Error(err)
	}

	res, err = client.PostForm(loginURL, vars)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res, http.StatusOK)
	if err != nil {
		t.Error(err)
	}

	content, contentType = testUpData("me.jpg", "single", t)
	res1, err = client.Post(upURL, contentType, content)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res1, http.StatusOK, "jpg")
	if err != nil {
		t.Error(err)
	}

	content, contentType = testUpData("me.jpg", "multi", t)
	res2, err = client.Post(upURL, contentType, content)
	defer res2.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res2.StatusCode != http.StatusOK {
		t.Errorf("Ecpected %d got %d", http.StatusOK, res2.StatusCode)
	}
	w.Reset()
	io.Copy(w, res2.Body)
	if !contains(w.String(), "jpg") {
		t.Errorf("Expected to save jpg file got %s", w.String())
	}

	bAb := rx.cfg.AccountsBucket
	rx.cfg.AccountsBucket = ""

	content, contentType = testUpData("me.jpg", "single", t)
	res3, err = client.Post(upURL, contentType, content)
	if err != nil {
		t.Error(err)
	}
	defer res3.Body.Close()
	if res3.StatusCode != http.StatusInternalServerError {
		t.Errorf("Ecpected %d got %d", http.StatusInternalServerError, res3.StatusCode)
	}
	w.Reset()
	io.Copy(w, res3.Body)
	if !contains(w.String(), "bucket") {
		t.Errorf("Expected to save jpg file got %s", w.String())
	}
	rx.cfg.AccountsBucket = bAb

	bAb = rx.cfg.ProfilePicField
	rx.cfg.ProfilePicField = ""

	content, contentType = testUpData("me.jpg", "single", t)
	res4, err = client.Post(upURL, contentType, content)
	defer res4.Body.Close()
	if err != nil {
		t.Error(err)
	}

	if res4.StatusCode != http.StatusInternalServerError {
		t.Errorf("Ecpected %d got %d", http.StatusInternalServerError, res4.StatusCode)
	}
	w.Reset()
	io.Copy(w, res4.Body)
	if !contains(w.String(), " no such file") {
		t.Errorf("Expected  %s to contain no such file", w.String())
	}
	rx.cfg.ProfilePicField = bAb
}

func TestRemixt_ServeImages(t *testing.T) {
	var (
		email      = "gernest@aurora.com"
		imagesPath = "/imgs"
		res        *http.Response
		err        error
		user       *User
		p          *Profile
	)
	ts, client, rx := testServer(t)
	defer ts.Close()

	user, err = GetUser(setDB(rx.db, rx.cfg.AccountsDB), rx.cfg.AccountsBucket, email)
	if err != nil {
		t.Error(err)
	}
	pdb := getProfileDatabase(rx.cfg.DBDir, user.UUID, rx.cfg.DBExtension)
	p, err = GetProfile(setDB(rx.db, pdb), rx.cfg.ProfilesBucket, user.UUID)
	if err != nil {
		t.Error(err)
	}
	if len(p.Photos) != 3 {
		t.Errorf("Expected 3 got %d", len(p.Photos))
	}

	query := url.Values{
		"iid": {p.Picture.ID},
		"pid": {p.ID},
	}
	imgURL := fmt.Sprintf("%s%s?%s", ts.URL, imagesPath, query.Encode())
	res, err = client.Get(imgURL)
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res.StatusCode)
	}

	// failure case
	vars := url.Values{
		"iid": {"bogus"},
		"pid": {p.Picture.UploadedBy},
	}
	res1, err := client.Get(fmt.Sprintf("%s%s?%s", ts.URL, imagesPath, vars.Encode()))
	defer res1.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res1.StatusCode != http.StatusNotFound {
		t.Errorf("Expected %d got %d", http.StatusNotFound, res1.StatusCode)
	}
}

func TestRemix_Logout(t *testing.T) {
	var (
		loginPath  = "/auth/login"
		logoutPath = "/auth/logout"
		email      = "gernest@aurora.com"
		pass       = "mamamia"
		err        error
		res, res1  *http.Response
		vars       url.Values
	)

	ts, client, _ := testServer(t)
	defer ts.Close()
	vars = url.Values{
		"email":    {email},
		"password": {pass},
	}

	loginURL := fmt.Sprintf("%s%s", ts.URL, loginPath)
	outURL := fmt.Sprintf("%s%s", ts.URL, logoutPath)

	res, err = client.PostForm(loginURL, vars)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res, http.StatusOK, "search")
	if err != nil {
		t.Error(err)
	}

	res1, err = client.Get(outURL)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res1, http.StatusOK, "search")
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestRemix_Profile(t *testing.T) {
	var (
		profilePath = "/profile"
		loginPath   = "/auth/login"
		pass        = "mamamia"
		birthDate   = "2 January, 1989"
		err         error
	)

	emails := []string{
		"gernest@aurora.mza",
		"gernest@aurora.tz",
		"gernest@aurora.tx",
	}

	ts, client, rx := testServer(t)
	defer ts.Close()

	// create user accounts and profiles, using the id's in pids global variables
	// and emails in the emsils slice. The id, email pairs correspond to the two
	// slice's index
	for k, v := range pids {
		usr := &User{EmailAddress: emails[k], UUID: v}
		ps, err := hashPassword(pass)
		if err != nil {
			t.Error(err)
		}
		usr.Pass = ps
		err = CreateAccount(setDB(rx.db, rx.cfg.AccountsDB), usr, rx.cfg.AccountsBucket)
		if err != nil {
			t.Error(err)
		}
		pdbStr := getProfileDatabase(rx.cfg.DBDir, usr.UUID, rx.cfg.DBExtension)
		pdb := setDB(rx.db, pdbStr)
		p := &Profile{ID: usr.UUID}
		err = CreateProfile(pdb, p, rx.cfg.ProfilesBucket)
		if err != nil {
			t.Error(err)
		}

	}

	for _, v := range pids {

		// A correct profile url query, this is for viewing a single profile only
		vars := url.Values{
			"id":   {v},
			"view": {"true"},
			"all":  {"false"},
		}

		// A wrong profile url query, notice that the id field is not a correct
		// uuid string and also there aint no such profiles in the database.
		// This also is for viewing a single profile
		vars2 := url.Values{
			"id":   {v + "shit"},
			"view": {"true"},
			"all":  {"false"},
		}

		// case a valid profile query, and the request is standard http.
		res, err := httpGet(client, fmt.Sprintf("%s/profile?%s", ts.URL, vars.Encode()))
		if err != nil {
			t.Error(err)
		}
		err = checkResponse(res, http.StatusOK)
		if err != nil {
			t.Error(err)
		}

		// case wrong profile url query, to be precise, the id is wrong that is it is not
		// a valid uuid and no any profile matches. The request is standard http.
		res0, err := httpGet(client, fmt.Sprintf("%s/profile?%s", ts.URL, vars2.Encode()))
		if err != nil {
			t.Error(err)
		}
		err = checkResponse(res0, http.StatusNotFound, "shit not found")
		if err != nil {
			t.Error(err)
		}

		// case a valid profile query, and the request is standard ajax.
		res1, err := httpGetAjax(client, fmt.Sprintf("%s/profile?%s", ts.URL, vars.Encode()))
		if err != nil {
			t.Error(err)
		}
		err = checkResponse(res1, http.StatusOK, v)
		if err != nil {
			t.Error(err)
		}
		// case wrong profile url query, to be precise, the id is wrong that is it is not
		// a valid uuid and no any profile matches. The request is standard  ajax.
		res2, err := httpGetAjax(client, fmt.Sprintf("%s/profile?%s", ts.URL, vars2.Encode()))
		if err != nil {
			t.Error(err)
		}
		err = checkResponse(res2, http.StatusNotFound, errNotFound.Error())
		if err != nil {
			t.Error(err)
		}
	}

	// A correct profile url query for viewing all profiles
	vars := url.Values{
		"view": {"true"},
		"all":  {"true"},
	}

	// case viewing all profiles via standard http
	res3, err := httpGet(client, fmt.Sprintf("%s/profile?%s", ts.URL, vars.Encode()))
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res3, http.StatusOK)
	if err != nil {
		t.Error(err)
	}

	// case viewing all profiles via ajax
	res4, err := httpGetAjax(client, fmt.Sprintf("%s/profile?%s", ts.URL, vars.Encode()))
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res4, http.StatusOK, pids[0])
	if err != nil {
		t.Error(err)
	}

	// Inorder to test what if the hadler woks fine when an internal server error
	// pccur. We set the accounts bucket to "", note that this hsould be restored
	// after this test finish inorder for other tests to work properly.
	//
	// All handlers reiles on the rx.cfg object heavily.
	bAcc := rx.cfg.AccountsBucket
	rx.cfg.AccountsBucket = ""

	// case an ajax request
	res5, err := httpGetAjax(client, fmt.Sprintf("%s/profile?%s", ts.URL, vars.Encode()))
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res5, http.StatusNotFound, errNotFound.Error())
	if err != nil {
		t.Error(err)
	}

	// case a standard http request
	res6, err := httpGet(client, fmt.Sprintf("%s/profile?%s", ts.URL, vars.Encode()))
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res6, http.StatusNotFound, "shit not found")
	if err != nil {
		t.Error(err)
	}

	// Restore the accounts bucket config value
	rx.cfg.AccountsBucket = bAcc

	profileForm := url.Values{
		"city":    {"mwanza"},
		"country": {"Tanzania"},
	}
	vars = url.Values{
		"u":  {"true"},
		"id": {pids[0]},
	}

	// case posting a valid form but the user is not logged in, the request is a standard http one.
	res7, err := client.PostForm(fmt.Sprintf("%s%s?%s", ts.URL, profilePath, vars.Encode()), profileForm)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res7, http.StatusOK, "login-form")
	if err != nil {
		t.Error(err)
	}

	// case posting a valid form but the user is not logged in, the request is ajax.
	res8, err := httpPostAjax(client, fmt.Sprintf("%s/profile?%s", ts.URL, vars.Encode()), strings.NewReader(profileForm.Encode()))
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res8, http.StatusForbidden, errForbidden.Error())
	if err != nil {
		t.Error(err)
	}
	// login and create a session for user with pids[0]
	varsLogin := url.Values{
		"email":    {emails[0]},
		"password": {pass},
	}
	res9, err := client.PostForm(fmt.Sprintf("%s%s", ts.URL, loginPath), varsLogin)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res9, http.StatusOK, "search")
	if err != nil {
		t.Error(err)
	}
	vars = url.Values{
		"u":  {"true"},
		"id": {pids[1]},
	}

	// case posting a valid form and the user is  logged in, the request is a standard http one.
	// The loggedIn user ID is defferent from the id provided by the url.
	res10, err := client.PostForm(fmt.Sprintf("%s%s?%s", ts.URL, profilePath, vars.Encode()), profileForm)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res10, http.StatusInternalServerError, "forbidden")
	if err != nil {
		t.Error(err)
	}

	varsTrue := url.Values{
		"u":  {"true"},
		"id": {pids[0]},
	}

	// The profile url which has the id query match logged user id
	loggedUsrURL := fmt.Sprintf("%s%s?%s", ts.URL, profilePath, varsTrue.Encode())

	// case posting an  invalid form but the user is logged in, the request is a standard http one.
	profileForm2 := url.Values{
		"city":       {"mwanza"},
		"country":    {"Tanzania"},
		"age":        {"12"},
		"birth_date": {birthDateFormat},
	}
	res11, err := client.PostForm(loggedUsrURL, profileForm2)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res11, http.StatusOK, "umri unatakiwa uwe zaidi ya miaka 18")
	if err != nil {
		t.Error(err)
	}
	// case posting a valid form, the user is logged in and the request is standard http
	profileForm3 := url.Values{
		"first_name": {"geofrey"},
		"last_name":  {"ernest"},
		"gender":     {"1"},
		"street":     {"kilimahewa"},
		"city":       {"mwanza"},
		"country":    {"Tanzania"},
		"age":        {"19"},
		"birth_date": {birthDate},
	}
	res12, err := client.PostForm(loggedUsrURL, profileForm3)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res12, http.StatusOK, birthDate)
	if err != nil {
		t.Errorf("checking response %v", err)
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
		AccountsBucket:      "accounts",
		SessionName:         "aurora",
		LoginRedirect:       "/",
		DBDir:               "fixture",
		DBExtension:         ".bdb",
		AccountsDB:          "fixture/accounts.bdb",
		ProfilesBucket:      "profiles",
		SessionsDB:          "fixture/sessions.bdb",
		SessionsBucket:      "sessions",
		ProfilePicField:     "profile",
		PhotosField:         "photos",
		TemplatesDir:        "templates",
		TemplatesExtensions: []string{".tmpl", ".html", ".tpl"},
		SessMaxAge:          30,
		SessionPath:         "/",
		MessagesBucket:      "messages",
	}
	rx := NewRemix(cfg)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{Jar: jar}
	ts := httptest.NewServer(rx.Routes())
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

func httpGet(client *http.Client, url string) (*http.Response, error) {
	h := make(http.Header)
	return httpCall(client, "GET", url, h, nil)
}

func httpGetAjax(client *http.Client, url string) (*http.Response, error) {
	h := make(http.Header)
	h.Set("X-Requested-With", "XMLHttpRequest")
	return httpCall(client, "GET", url, h, nil)
}

func httpPostAjax(client *http.Client, url string, body io.Reader) (*http.Response, error) {
	h := make(http.Header)
	h.Set("X-Requested-With", "XMLHttpRequest")
	return httpCall(client, "POST", url, h, body)
}

func httpCall(client *http.Client, method, url string, header http.Header, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for k, vs := range header {
		req.Header[k] = vs
	}
	return client.Do(req)
}
func checkResponse(res *http.Response, status int, contain ...string) error {
	defer res.Body.Close()
	var err listErr
	w := &bytes.Buffer{}
	io.Copy(w, res.Body)
	if res.StatusCode != status {
		err = append(err, fmt.Errorf("Expected %d got %d \n", status, res.StatusCode))
	}

	if len(contain) > 0 {
		if !contains(w.String(), contain[0]) {
			err = append(err, fmt.Errorf("Expected %s to contain %s", w.String(), contain[0]))
		}
	}
	if len(err) > 0 {
		return err
	}
	return nil
}
