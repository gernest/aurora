package aurora

import (
	"net/http"
	"time"

	"github.com/gernest/golem"
)

const (
	mainRoom string = "aurora"

	// events
	sendEvt       string = "send"
	receiveEvt    string = "receive"
	sendFailedEvt string = "sendFailed"

	// message buckets
	ouboxBucket string = "outbox"
	inboxBucket string = "inbox"
	draftBucket string = "drafts"

	// message allerts
	alertSendSuccess string = "sendSuccess"
	alertSendFailed  string = "sendFailled"
	alertInbox       string = "messageInbox"
)

// MSG this is the base message exchanged between users
type MSG struct {
	ID          string    `json:"id"`
	SenderID    string    `json:"sender_id"`
	RecipientID string    `json:"recepient_id"`
	Text        string    `json:"text"`
	SentAt      time.Time `json:"sent_at"`
	ReceivedAt  time.Time `json:"received_at"`
	Status      int       `json:"status"`
}

// InfoMSG this is for sharing information across the messenger nodes
type InfoMSG struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Sender string `json:"sender"`
}

// Messenger the messanger from the gods
type Messenger struct {
	rx     *Remix
	rm     *golem.RoomManager
	route  *golem.Router
	online map[string]*Profile
}

// NewMessenger creates a new messenger
func NewMessenger(rx *Remix) *Messenger {
	return &Messenger{
		rx:     rx,
		rm:     golem.NewRoomManager(),
		route:  golem.NewRouter(),
		online: make(map[string]*Profile),
	}
}

func (m *Messenger) validateSession(w http.ResponseWriter, r *http.Request) bool {
	if ss, ok := m.rx.isInSession(r); ok && !ss.IsNew {
		return true
	}
	return false
}

// add  user to online user's list
func (m *Messenger) onConnect(conn *golem.Connection, r *http.Request) {
	if ss, ok := m.rx.isInSession(r); ok && !ss.IsNew {
		_, p, err := m.rx.getCurrentUserAndProfile(ss)
		if err == nil {
			if prof, ok := m.online[p.ID]; !ok && prof == nil {
				conn.UserID = p.ID
				conn.SetSendCallBack(m.callMeBack)
				m.rm.Join(mainRoom, conn)
				m.rm.Join(p.ID, conn)
			}
		}
	}
}

func (m *Messenger) onClose(conn *golem.Connection) {
	delete(m.online, conn.UserID)
}

func (m *Messenger) callMeBack(conn *golem.Connection, msg *golem.Message) *golem.Message {
	switch msg.GetEvent() {
	case sendEvt:
		switch data := msg.GetData().(type) {
		case *MSG:
			p := m.currentUser(conn)
			if p != nil {
				if p.ID != data.SenderID {
					data.Status = http.StatusBadRequest
					msg.SetEvent(alertSendFailed)
					msg.SetData(data)
					return msg
				}
				err := m.saveMsg(ouboxBucket, p.ID, data)
				if err != nil {
					data.Status = http.StatusInternalServerError
					return setMSG(alertSendFailed, data, msg)
				}
				if m.isOnline(data.RecipientID) {
					m.rm.Emit(data.RecipientID, receiveEvt, data)
					data.Status = http.StatusOK
					return setMSG(alertSendSuccess, data, msg)
				}
				err = m.saveMsg(inboxBucket, data.RecipientID, data)
				if err != nil {
					data.Status = http.StatusInternalServerError
					return setMSG(alertSendFailed, data, msg)
				}
				data.Status = http.StatusOK
				return setMSG(alertSendSuccess, data, msg)
			}
		}
	case receiveEvt:
		switch data := msg.GetData().(type) {
		case *MSG:
			p := m.currentUser(conn)
			if p != nil {
				if p.ID == data.RecipientID {
					err := m.saveMsg(inboxBucket, p.ID, data)
					if err != nil {
						if m.isOnline(data.SenderID) {
							// inform the sender that sending failed
							m.rm.Emit(data.SenderID, sendFailedEvt, data)
						}
						data.Status = http.StatusInternalServerError
						msg.SetData(data)
						return msg
					}
					data.Status = http.StatusOK
					msg.SetEvent(alertInbox)
					msg.SetData(data)
					return msg
				}
			}
		}
	case sendFailedEvt:
		switch data := msg.GetData().(type) {
		case *MSG:
			p := m.currentUser(conn)
			if p != nil {
				err := m.deletMsg(inboxBucket, p.ID, data)
				if err != nil {
					data.Status = http.StatusInternalServerError
					return setMSG("", data, msg)
				}
				err = m.saveMsg(draftBucket, p.ID, data)
				if err != nil {
					data.Status = http.StatusInternalServerError
					msg.SetData(data)
					return setMSG("", data, msg)
				}
				return setMSG(alertSendFailed, nil, msg)
			}

		}

	}
	return msg
}
func (m *Messenger) saveMsg(bucket string, profileID string, msg *MSG) error {
	if msg.ID == "" {
		msg.ID = getUUID()
	}
	pdb := getProfileDatabase(m.rx.cfg.DBDir, profileID, m.rx.cfg.DBExtension)
	mdb := setDB(m.rx.db, pdb)
	return marshalAndCreate(mdb, msg, bucket, msg.ID, m.rx.cfg.MessagesBucket)
}
func (m *Messenger) deletMsg(bucket, profileID string, msg *MSG) error {
	pdb := getProfileDatabase(m.rx.cfg.DBDir, profileID, m.rx.cfg.DBExtension)
	mdb := setDB(m.rx.db, pdb)
	mdb.Delete(bucket, msg.ID, m.rx.cfg.MessagesBucket)
	return mdb.Error
}
func (m *Messenger) currentUser(conn *golem.Connection) *Profile {
	pdb := getProfileDatabase(m.rx.cfg.DBDir, conn.UserID, m.rx.cfg.DBExtension)
	mdb := setDB(m.rx.db, pdb)
	p, err := GetProfile(mdb, m.rx.cfg.ProfilesBucket, conn.UserID)
	if err != nil {
		// log this
		return nil
	}
	return p
}

func (m *Messenger) isOnline(id string) bool {
	if p, ok := m.online[id]; ok && p != nil {
		return true
	}
	return false
}

func setMSG(evt string, data interface{}, msg *golem.Message) *golem.Message {
	if evt != "" {
		msg.SetEvent(evt)
	}
	if data != nil {
		msg.SetData(data)
	}
	return msg
}
func (m *Messenger) info(conn *golem.Connection, msg *InfoMSG) {
	m.rm.Emit(mainRoom, "info", msg)
}

func (m *Messenger) send(conn *golem.Connection, msg *MSG) {
	m.rm.Emit(msg.SenderID, sendEvt, msg)
}

// Handler handles websocket connections for messaging
func (m *Messenger) Handler() func(http.ResponseWriter, *http.Request) {
	m.route.OnHandshake(m.validateSession)
	m.route.OnConnect(m.onConnect)
	m.route.OnClose(m.onClose)
	m.route.On("info", m.info)
	m.route.On("send", m.send)
	return m.route.Handler()
}
