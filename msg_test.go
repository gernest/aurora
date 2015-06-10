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

	"github.com/gorilla/websocket"
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

	// Websocket route url.
	wsURL := fmt.Sprintf("ws://%s/msg", origin.Host)
	u, err := url.Parse(wsURL)
	if err != nil {
		t.Errorf("parsing wesocket url %v", err)
	}

	// There is no session yet, when
	aConn, err := net.Dial("tcp", origin.Host)
	if err != nil {
		t.Errorf("establishing a connection %v", err)
	}

	ws1, _, err := websocket.NewClient(aConn, u, make(http.Header), 1024, 1024)
	if err == nil {
		t.Error("expected an error got nil instead")
	}
	if ws1 != nil {
		t.Error("Expected nil")
		ws1.Close()
	}

	// Create a user and profile.  Login and start a new session. This should make
	// the handshake succeed.
	usr := &User{UUID: userID, EmailAddress: email}

	// NOTE: the user password is stored as a hash, so we have to generate the hash to
	// minic the user account.
	ps, err := hashPassword(pass)
	if err != nil {
		t.Error(err)
	}
	usr.Pass = ps
	err = CreateAccount(setDB(rx.db, rx.cfg.AccountsDB), usr, rx.cfg.AccountsBucket)
	if err != nil {
		t.Errorf("creating a new account %v", err)
	}

	// Create a new profile, based on the user we have created above.
	pdbStr := getProfileDatabase(rx.cfg.DBDir, usr.UUID, rx.cfg.DBExtension)
	pdb := setDB(rx.db, pdbStr)
	p := &Profile{ID: usr.UUID}
	err = CreateProfile(pdb, p, rx.cfg.ProfilesBucket)
	if err != nil {
		t.Errorf("creating profile: %v", err)
	}

	// login and create a session for the user we have just created.
	varsLogin := url.Values{"email": {email}, "password": {pass}}
	res9, err := client.PostForm(fmt.Sprintf("%s%s", ts.URL, loginPath), varsLogin)
	if err != nil {
		t.Error(err)
	}
	err = checkResponse(res9, http.StatusOK, "search")
	if err != nil {
		t.Error(err)
	}

	// Get the cookie data from the client, so that we can send the data in the header for
	// the following websocket request.
	h := make(http.Header)
	for _, cookie := range client.Jar.Cookies(origin) {
		h.Set("Cookie", cookie.String())
	}

	// A bad request should fail, tmsg is the message to be send for a chat.
	tmsg := &MSG{SenderID: userID, Text: "hellp gernest", RecipientID: usr.UUID}
	d, err := marshalAndPach("send", tmsg)
	if err != nil {
		t.Errorf("marshaling and packing %v", err)
	}
	nConn, err := net.Dial("tcp", origin.Host)
	if err != nil {
		t.Errorf("establishing a connection %v", err)
	}
	ws3, _, err := websocket.NewClient(nConn, u, h, 1024, 1024)
	if err != nil {
		t.Errorf("extablishing websocket connection %v", err)
	}
	defer ws3.Close()

	// Set a one second deadline to the connection
	setDeadline(t, ws3)

	err = ws3.WriteMessage(websocket.TextMessage, d)
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

// Marshalls and pack the message( dPtr ) into a protocl that is used to transfer
// messages.
func marshalAndPach(name string, dPtr interface{}) ([]byte, error) {
	var protocolSeperator = " "
	if data, err := json.Marshal(dPtr); err == nil {
		result := []byte(name + protocolSeperator)
		return append(result, data...), nil
	} else {
		return nil, err
	}
}

// Decodes received message from the chat server.
func unpackMSG(data []byte) (string, []byte, error) {
	protocolSeperator := " "
	result := strings.SplitN(string(data), protocolSeperator, 2)
	if len(result) != 2 {
		return "", nil, errors.New("Unable to extract event name from data.")
	}
	return result[0], []byte(result[1]), nil
}

// sets deadline to a websocket connection
func setDeadline(t *testing.T, ws *websocket.Conn) {
	err := ws.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("set write deadline %v", err)
	}
	err = ws.SetReadDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("set read deadline %v", err)
	}
}
