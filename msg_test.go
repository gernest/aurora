package aurora

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	gs "github.com/gorilla/websocket"
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
		t.Errorf("creating profile: %v", err)
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

	// try sending a bad request
	tmsg := &MSG{
		SenderID:    userID,
		Text:        "hellp gernest",
		RecipientID: usr.UUID,
	}
	d, err := marshalAndPach("send", tmsg)
	if err != nil {
		t.Errorf("marshaling and packing %v", err)
	}

	nConn, err := net.Dial("tcp", origin.Host)
	if err != nil {
		t.Errorf("establishing a connection %v", err)
	}

	u, err := url.Parse(wsURL)
	if err != nil {
		t.Errorf("parsing wesocket url %v", err)
	}
	ws3, _, err := gs.NewClient(nConn, u, h, 1024, 1024)
	if err != nil {
		t.Errorf("extablishing websocket connection %v", err)
	}
	defer ws3.Close()
	setDeadline(t, ws3)
	err = ws3.WriteMessage(gs.TextMessage, d)
	if err != nil {
		t.Errorf("writing message %v", err)
	}
	_, rs, err := ws3.ReadMessage()
	if err != nil {
		t.Errorf("reading message %v", err)
	}
	evt, dmsg, err := unpackMSG(rs)
	if err != nil {
		t.Errorf("unpacking read message %v \n %s", err, string(rs))

	}
	if evt != alertSendSuccess {
		t.Errorf("Expected %s got %s", alertSendSuccess, evt)
	}
	if !contains(string(dmsg), tmsg.Text) {
		t.Errorf("Expected %s to contain %s", string(rs), tmsg.Text)
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

func unpackMSG(data []byte) (string, []byte, error) {
	protocolSeperator := " "
	result := strings.SplitN(string(data), protocolSeperator, 2)
	if len(result) != 2 {
		return "", nil, errors.New("Unable to extract event name from data.")
	}
	return result[0], []byte(result[1]), nil
}

func setDeadline(t *testing.T, ws *gs.Conn) {
	err := ws.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("set write deadline %v", err)
	}
	err = ws.SetReadDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("set read deadline %v", err)
	}
}
