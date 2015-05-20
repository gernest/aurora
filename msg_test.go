package aurora

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func TestMessenger(t *testing.T) {
	var (
		email     = "wesuckssoomuch@aurora.com"
		pass      = "mamamia"
		loginPath = "/auth/login"
		userID    = "37c37153-089e-4c19-466e-2f467ac07c1e"
	)

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

	// create a user and profile
	usr := &User{
		UUID:         userID,
		EmailAddress: email,
	}
	ps, err := hashPassword(pass)
	if err != nil {
		t.Error(err)
	}
	usr.Pass = ps
	err = CreateAccount(setDB(rx.db, rx.cfg.AccountsDB), usr, rx.cfg.AccountsBucket)
	if err != nil {
		t.Errorf("creating a new account %v", err)
	}
	pdbStr := getProfileDatabase(rx.cfg.DBDir, usr.UUID, rx.cfg.DBExtension)
	pdb := setDB(rx.db, pdbStr)
	p := &Profile{ID: usr.UUID}
	err = CreateProfile(pdb, p, rx.cfg.ProfilesBucket)
	if err != nil {
		t.Error(err)
	}

	// There is no session yet, it should fail to validate
	ws1, err := websocket.DialConfig(cfg)
	if err == nil {
		t.Error("Expected an error")
	}
	if ws1 != nil {
		t.Error("Expected nil")
		ws1.Close()
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

	ws2, err := websocket.DialConfig(cfg2)
	if err != nil {
		t.Error(err)
	}
	defer ws2.Close()

	// try sending a bad request
	tmsg := &MSG{
		SenderID: getUUID(),
		Text:     "hellp gernest",
	}
	d, err := marshalAndPach("send", tmsg)
	if err != nil {
		t.Errorf("marshaling and packing %v")
	}

	err = ws2.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("set write deadline %v", err)
	}
	_, err = ws2.Write(d)
	if err != nil {
		t.Errorf("writing message %v", err)
	}
}

func marshalAndPach(name string, dPtr interface{}) ([]byte, error) {
	var protocolSeperator = " "
	if data, err := json.Marshal(dPtr); err == nil {
		result := []byte(name + protocolSeperator)
		return append(result, data...), nil
	} else {
		return nil, err
	}
}

func unpackMSG(data []byte) (string, interface{}, error) {
	var protocolSeperator = " "
	result := strings.SplitN(string(data), protocolSeperator, 2)
	if len(result) != 2 {
		return "", nil, errors.New("Unable to extract event name from data.")
	}
	return result[0], []byte(result[1]), nil
}
