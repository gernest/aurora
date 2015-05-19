package aurora

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/net/websocket"
)

func TestMessenger(t *testing.T) {
	var email = "wesucks@aurora.com"
	var pass = "mamamia"
	var loginPath = "/auth/login"

	ts, client, rx := testServer(t)
	defer ts.Close()
	if rx != nil {
	}
	if client != nil {
	}
	origin, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err)
	}
	wsURL := fmt.Sprintf("ws://%s/msg", origin.Host)
	cfg, err := websocket.NewConfig(wsURL, ts.URL)
	if err != nil {
		t.Error(err)
	}

	// There is no session yet, it should fail to validate
	ws, err := websocket.DialConfig(cfg)
	if err == nil {
		t.Error("Expected an error")
	}
	if ws != nil {
		t.Error("Expected nil")
	}

	// create a user and profile
	usr := &User{
		UUID:         getUUID(),
		EmailAddress: email,
	}
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
	// login and create a session for user with pids[0]
	varsLogin := url.Values{
		"email":    {email},
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
	h := make(http.Header)
	for _, cookie := range client.Jar.Cookies(origin) {
		h.Set("Cookie", cookie.String())
	}
	cfg2, err := websocket.NewConfig(wsURL, ts.URL)
	if err != nil {
		t.Error(err)
	}
	cfg2.Header = h

	ws, err = websocket.DialConfig(cfg2)
	if err != nil {
		t.Error(err)
	}
	ws.Close()
}
