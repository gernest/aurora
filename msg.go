package aurora

import (
	"net/http"
	"time"

	"github.com/muesli/cache2go"

	"github.com/gernest/golem"
)

const (
	mainRoom = "aurora"

	// events
	sendEvt       = "send"
	receiveEvt    = "receive"
	sendFailedEvt = "sendFailed"
	ignore        = "ignore"

	// message buckets
	outboxBucket = "outbox"
	inboxBucket  = "inbox"
	draftBucket  = "drafts"

	// message allerts
	alertSendSuccess = "sendSuccess"
	alertSendFailed  = "sendFailled"
	alertInbox       = "messageInbox"

	// cache
	onlineCache = "online"
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
	online *cache2go.CacheTable
}

// NewMessenger creates a new messenger
func NewMessenger(rx *Remix) *Messenger {
	return &Messenger{
		rx:     rx,
		rm:     golem.NewRoomManager(),
		route:  golem.NewRouter(),
		online: cache2go.Cache(onlineCache),
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
			conn.UserID = p.ID
			conn.SetSendCallBack(m.callMeBack)
			m.rm.Join(mainRoom, conn)
			m.rm.Join(p.ID, conn)
			m.online.Add(p.ID, 0, p)
		}
	}
}

func (m *Messenger) callMeBack(conn *golem.Connection, msg *golem.Message) *golem.Message {
	p := m.currentUser(conn)
	switch msg.GetEvent() {
	case sendEvt:
		switch data := msg.GetData().(type) {
		case *MSG:
			if p != nil {
				if p.ID == data.SenderID {

					data.SentAt = time.Now()
					err := m.saveMsg(outboxBucket, p.ID, data)
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
		}
	case receiveEvt:
		switch data := msg.GetData().(type) {
		case *MSG:
			if p != nil {
				if p.ID == data.RecipientID {
					data.ReceivedAt = time.Now()
					err := m.saveMsg(inboxBucket, p.ID, data)
					if err != nil {
						msg.SetEvent(ignore)
						if m.isOnline(data.SenderID) {
							m.rm.Emit(data.SenderID, sendFailedEvt, data)
							return msg
						}
						err = m.moveTo(draftBucket, outboxBucket, data.SenderID, data.ID)
						if err != nil {
							// TODO: log this?
						}
						return msg
					}
					data.Status = http.StatusOK
					return setMSG(alertInbox, data, msg)
				}
			}
		}
	case sendFailedEvt:
		switch data := msg.GetData().(type) {
		case *MSG:
			if p != nil && data.SenderID == p.ID {
				err := m.moveTo(draftBucket, outboxBucket, p.ID, data.ID)
				if err != nil {
					// TODO: log this?
				}
				return setMSG(alertSendFailed, nil, msg)
			}

		}

	}
	return msg
}

// persist a message
func (m *Messenger) saveMsg(bucket string, profileID string, msg *MSG) error {
	if msg.ID == "" {
		msg.ID = getUUID()
	}
	pdb := getProfileDatabase(m.rx.cfg.DBDir, profileID, m.rx.cfg.DBExtension)
	mdb := setDB(m.rx.db, pdb)
	return marshalAndCreate(mdb, msg, bucket, msg.ID, m.rx.cfg.MessagesBucket)
}

// deletes a message
func (m *Messenger) deletMsg(bucket, profileID string, msg *MSG) error {
	pdb := getProfileDatabase(m.rx.cfg.DBDir, profileID, m.rx.cfg.DBExtension)
	mdb := setDB(m.rx.db, pdb)
	mdb.Delete(bucket, msg.ID, m.rx.cfg.MessagesBucket)
	return mdb.Error
}

// moves message data from one bucket to another.
func (m *Messenger) moveTo(dest, src, profileID, msgID string) error {
	pdb := getProfileDatabase(m.rx.cfg.DBDir, profileID, m.rx.cfg.DBExtension)
	mdb := setDB(m.rx.db, pdb)
	d := mdb.Get(src, msgID, m.rx.cfg.MessagesBucket)
	if d.Error != nil {
		return d.Error
	}
	s := mdb.Create(dest, msgID, d.Data, m.rx.cfg.MessagesBucket)
	if s.Error != nil {
		return s.Error
	}
	return mdb.Delete(src, msgID, m.rx.cfg.MessagesBucket).Error
}

// gets the user's profile of a given websocket connection.
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

// sets the event and data attributes of a givenMSG.
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

func (m *Messenger) onClose(conn *golem.Connection) {
	m.online.Delete(conn.UserID)
	m.rm.Leave(conn.UserID, conn)
}

func (m *Messenger) isOnline(key string) bool {
	return m.online.Exists(key)
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
