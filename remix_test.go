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
		w   = &bytes.Buffer{}
		err error
		res *http.Response
	)
	ts, client, _ := testServer(t)
	defer ts.Close()

	res, err = client.Get(ts.URL)
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res.StatusCode)
	}
	io.Copy(w, res.Body)
	if !contains(w.String(), "prove it yourself") {
		t.Error("Expected InSession not to be pset")
	}
}

func TestRemix_Register(t *testing.T) {
	var (
		w                            = &bytes.Buffer{}
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
	defer res1.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res1.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res1.StatusCode)
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
	defer res2.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res2.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res2.StatusCode)
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
	defer res3.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res3.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected %d got %d", http.StatusFound, res3.StatusCode)
	}
	rx.cfg.AccountsBucket = "accounts" //Restore our config

	// case everything is ok
	res4, err = client.PostForm(registerURL, vars)
	defer res4.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res4.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res4.StatusCode)
	}
	io.Copy(w, res4.Body)
	if !contains(w.String(), "search") {
		t.Error("Expected InSession to be set")
	}

	// case session is not new it should redirects to login path
	res5, err = client.PostForm(registerURL, vars)
	defer res5.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res5.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res5.StatusCode)
	}
	w.Reset() // reuse this buffer
	io.Copy(w, res5.Body)
	if !contains(w.String(), "/auth/logout") {
		t.Errorf("Expected login form got %s", w.String())
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
		w                     = &bytes.Buffer{}
		email                 = "gernest@aurora.com"
		loginPath             = "/auth/login"
		pass                  = "mamamia"
		err                   error
		res, res1, res2, res3 *http.Response
		vars                  url.Values
	)

	ts, client, _ := testServer(t)
	defer ts.Close()
	loginURL := fmt.Sprintf("%s%s", ts.URL, loginPath)

	// get the login form
	res, err = client.Get(loginURL)
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res.StatusCode)
	}

	// invalid form
	vars = url.Values{
		"email":    {"bogus"},
		"password": {"myass"},
	}
	res1, err = client.PostForm(loginURL, vars)
	defer res1.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res1.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res1.StatusCode)
	}
	io.Copy(w, res1.Body)
	if !contains(w.String(), "login-form") {
		t.Error("Expected login form")
	}

	// case no such user but valid form
	vars = url.Values{
		"email":    {"gernesti@aurora.com"},
		"password": {"heydollringadongdillo"},
	}
	res2, err = client.PostForm(loginURL, vars)
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
	vars = url.Values{
		"email":    {"gernest@aurora.com"},
		"password": {"heydollringadongdilloo"},
	}
	res3, err = client.PostForm(loginURL, vars)
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
	vars = url.Values{
		"email":    {email},
		"password": {pass},
	}
	res3, err = client.PostForm(loginURL, vars)
	defer res3.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res3.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res3.StatusCode)
	}
	w.Reset()
	io.Copy(w, res3.Body)

	if !contains(w.String(), "search") {
		t.Error("Expected InSession to be set")
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
	defer res0.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res0.StatusCode != http.StatusForbidden {
		t.Errorf("Ecpected %d got %d", http.StatusForbidden, res0.StatusCode)
	}
	io.Copy(w, res0.Body)
	if !contains(w.String(), errForbidden.Error()) {
		t.Errorf("Expected to be forbidden got %s", w.String())
	}

	res, err = client.PostForm(loginURL, vars)
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Ecpected %d got %d", http.StatusOK, res.StatusCode)
	}
	content, contentType = testUpData("me.jpg", "single", t)
	res1, err = client.Post(upURL, contentType, content)
	defer res1.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res1.StatusCode != http.StatusOK {
		t.Errorf("Ecpected %d got %d", http.StatusOK, res1.StatusCode)
	}
	io.Copy(w, res1.Body)
	if !contains(w.String(), "jpg") {
		t.Errorf("Expected to save jpg file got %s", w.String())
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
	imgURL := fmt.Sprintf("%s%s?%s", ts.URL, imagesPath, p.Picture.Query)

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
		w          = &bytes.Buffer{}
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
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Ecpected %d got %d", http.StatusOK, res.StatusCode)
	}
	io.Copy(w, res.Body)
	if !contains(w.String(), "search") {
		t.Error("Expected InSession to be set")
	}
	res1, err = client.Get(outURL)
	defer res1.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res1.StatusCode != http.StatusOK {
		t.Errorf("Ecpected %d got %d", http.StatusOK, res1.StatusCode)
	}
	w.Reset()
	io.Copy(w, res1.Body)
	if contains(w.String(), "search") {
		t.Error("Expected not to be in session")
	}
}

func TestRemix_Profile(t *testing.T) {
	var (
		profilePath                  = "/profile"
		loginPath                    = "/auth/login"
		pass                         = "mamamia"
		birthDate                    = "14 Apr 97 13:33 EAT"
		err                          error
		req, req1, req2, req3        *http.Request
		req4, req5, req6, req7, req8 *http.Request
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

	// Get the routes, this is so as to test the parts that don't require the user
	// to be loggen in, so using the client will be overkill.
	h := rx.Routes()

	/*
		=================================================================
		GET  requests
		=================================================================
	*/
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
		req, err = http.NewRequest("GET", fmt.Sprintf("/profile?%s", vars.Encode()), nil)
		if err != nil {
			t.Error(err)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, w.Code)
		}

		//// well till when I fix the templates this should work too
		//if !contains(w.Body.String(), v) {
		//	t.Errorf("Expected %s to contain %s", w.Body.String(), v)
		//}

		// case wrong profile url query, to be precise, the id is wrong that is it is not
		// a valid uuid and no any profile matches. The request is standard http.
		req1, err = http.NewRequest("GET", fmt.Sprintf("/profile?%s", vars2.Encode()), nil)
		if err != nil {
			t.Error(err)
		}
		w = httptest.NewRecorder()
		h.ServeHTTP(w, req1)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected %d got %d", http.StatusNotFound, w.Code)
		}
		if !contains(w.Body.String(), "shit not found") {
			t.Error("Expected a 404 custom template to be rendered")
		}

		// case a valid profile query, and the request is standard ajax.
		req2, err = http.NewRequest("GET", fmt.Sprintf("/profile?%s", vars.Encode()), nil)
		if err != nil {
			t.Error(err)
		}
		req2.Header.Set("X-Requested-With", "XMLHttpRequest")
		w = httptest.NewRecorder()
		h.ServeHTTP(w, req2)
		if w.Code != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, w.Code)
		}
		if !contains(w.Body.String(), v) {
			t.Errorf("Expected %s to contain %s", w.Body.String(), v)
		}

		// case wrong profile url query, to be precise, the id is wrong that is it is not
		// a valid uuid and no any profile matches. The request is standard  ajax.
		req3, err = http.NewRequest("GET", fmt.Sprintf("/profile?%s", vars2.Encode()), nil)
		if err != nil {
			t.Error(err)
		}
		req3.Header.Set("X-Requested-With", "XMLHttpRequest")
		w = httptest.NewRecorder()
		h.ServeHTTP(w, req3)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected %d got %d", http.StatusNotFound, w.Code)
		}
		if !contains(w.Body.String(), errNotFound.Error()) {
			t.Errorf("Expected %s to contain %s", w.Body.String(), errNotFound.Error())
		}
	}

	// A correct profile url query for viewing all profiles
	vars := url.Values{
		"view": {"true"},
		"all":  {"true"},
	}

	// case viewing all profiles via standard http
	req4, err = http.NewRequest("GET", fmt.Sprintf("/profile?%s", vars.Encode()), nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req4)
	if w.Code != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, w.Code)
	}

	//// well till when I fix the templates this should work too
	//if !contains(w.Body.String(), v) {
	//	t.Errorf("Expected %s to contain %s", w.Body.String(), v)
	//}

	// case viewing all profiles via ajax
	req5, err = http.NewRequest("GET", fmt.Sprintf("/profile?%s", vars.Encode()), nil)
	if err != nil {
		t.Error(err)
	}
	req5.Header.Set("X-Requested-With", "XMLHttpRequest")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req5)
	if w.Code != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, w.Code)
	}
	if !contains(w.Body.String(), pids[0]) {
		t.Errorf("Expected %s to contain %s", w.Body.String(), pids[0])
	}

	// Inorder to test what if the hadler woks fine when an internal server error
	// pccur. We set the accounts bucket to "", note that this hsould be restored
	// after this test finish inorder for other tests to work properly.
	//
	// All handlers reiles on the rx.cfg object heavily.
	bAcc := rx.cfg.AccountsBucket
	rx.cfg.AccountsBucket = ""
	req6, err = http.NewRequest("GET", fmt.Sprintf("/profile?%s", vars.Encode()), nil)
	if err != nil {
		t.Error(err)
	}

	// case a standard ajax rewuest.
	req6.Header.Set("X-Requested-With", "XMLHttpRequest")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req6)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected %d got %d", http.StatusNotFound, w.Code)
	}
	if !contains(w.Body.String(), errNotFound.Error()) {
		t.Errorf("Expected %s to contain %s", w.Body.String(), errNotFound.Error())
	}

	// case a standard http request
	req7, err = http.NewRequest("GET", fmt.Sprintf("/profile?%s", vars.Encode()), nil)
	if err != nil {
		t.Error(err)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req7)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected %d got %d", http.StatusNotFound, w.Code)
	}
	if !contains(w.Body.String(), "shit not found") {
		t.Errorf("Expected 404 page got %s", w.Body.String())
	}

	// Restore the accounts bucket config value
	rx.cfg.AccountsBucket = bAcc

	/*
		===============================================================
		POST requests
		===============================================================
	*/

	profileForm := url.Values{
		"city":    {"mwanza"},
		"country": {"Tanzania"},
	}
	vars = url.Values{
		"u":  {"true"},
		"id": {pids[0]},
	}

	// case posting a valid form but the user is not logged in, the request is a standard http one.
	res, err := client.PostForm(fmt.Sprintf("%s%s?%s", ts.URL, profilePath, vars.Encode()), profileForm)
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res.StatusCode)
	}
	buf := &bytes.Buffer{}
	io.Copy(buf, res.Body)
	if !contains(buf.String(), "login-form") {
		t.Errorf("Expected to redirect to login path got %s", buf.String())
	}

	// case posting a valid form but the user is not logged in, the request is ajax.
	req8, err = http.NewRequest("POST", fmt.Sprintf("/profile?%s", vars.Encode()), strings.NewReader(profileForm.Encode()))
	if err != nil {
		t.Error(err)
	}
	req8.Header.Set("X-Requested-With", "XMLHttpRequest")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req8)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected %d got %d", http.StatusForbidden, w.Code)
	}
	if !contains(w.Body.String(), errForbidden.Error()) {
		t.Errorf("Expected %s to contain %s", w.Body.String(), errForbidden.Error())
	}

	// login and create a session for user with pids[0]
	varsLogin := url.Values{
		"email":    {emails[0]},
		"password": {pass},
	}
	res1, err := client.PostForm(fmt.Sprintf("%s%s", ts.URL, loginPath), varsLogin)
	defer res1.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res1.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res1.StatusCode)
	}
	buf.Reset()
	io.Copy(buf, res1.Body)
	if !contains(buf.String(), "search") {
		t.Errorf("Expected to be in session got %s", buf.String())
	}
	vars = url.Values{
		"u":  {"true"},
		"id": {pids[1]},
	}
	varsTrue := url.Values{
		"u":  {"true"},
		"id": {pids[0]},
	}

	// The profile url which has the id query match logged user id
	loggedUsrURL := fmt.Sprintf("%s%s?%s", ts.URL, profilePath, varsTrue.Encode())

	// case posting a valid form and the user is  logged in, the request is a standard http one.
	// The loggedIn user ID is defferent from the id provided by the url.
	res2, err := client.PostForm(fmt.Sprintf("%s%s?%s", ts.URL, profilePath, vars.Encode()), profileForm)
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res2.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected %d got %d", http.StatusInternalServerError, res.StatusCode)
	}
	buf.Reset()
	io.Copy(buf, res2.Body)
	if !contains(buf.String(), "forbidden") {
		t.Errorf("Expected to render 403 template got %s", buf.String())
	}

	// case posting an  invalid form but the user is logged in, the request is a standard http one.
	profileForm2 := url.Values{
		"city":       {"mwanza"},
		"country":    {"Tanzania"},
		"age":        {"12"},
		"birth_date": {birthDate},
	}
	res3, err := client.PostForm(loggedUsrURL, profileForm2)
	defer res3.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res3.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res3.StatusCode)
	}
	buf.Reset()
	io.Copy(buf, res3.Body)
	if !contains(buf.String(), "profile-home") {
		t.Errorf("Expected to have rendered profile/home got %s", buf.String())
	}

	// case posting a valid form, the user is logged in and the request is standard http
	profileForm3 := url.Values{
		"city":       {"mwanza"},
		"country":    {"Tanzania"},
		"age":        {"19"},
		"birth_date": {birthDate},
	}
	res4, err := client.PostForm(loggedUsrURL, profileForm3)
	defer res4.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if res4.StatusCode != http.StatusOK {
		t.Errorf("Expected %d got %d", http.StatusOK, res4.StatusCode)
	}
	buf.Reset()
	io.Copy(buf, res4.Body)
	if !contains(buf.String(), "profile-home") {
		t.Errorf("Expected to have rendered profile/home got %s", buf.String())
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
