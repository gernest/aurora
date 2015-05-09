package aurora

import (
	"log"
	"net/http"

	"github.com/gernest/nutz"
	"github.com/gorilla/sessions"

	"github.com/gernest/render"
)

// Remix all the fun is here
type Remix struct {
	db    nutz.Storage
	sess  *Session
	rendr *render.Render
	cfg   *RemixConfig
}

// RemixConfig contain configuration values for Remix
type RemixConfig struct {
	AppName        string `json:"name"`
	AppURL         string `json:"url"`
	CdnMode        bool   `json:"cdn_mode"`
	RunMode        string `json:"run_mode"`
	AppTitle       string `json:"title"`
	AppDescription string `json:"description"`
	SessionName    string `json:"session_name"`
	AccountsBucket string `json:"accounts_bucket"`
	AccountsDB     string

	// The path to point to when login is success
	LoginRedirect string `json:"login_redirect"`

	DBDir          string `json:"database_dir"`
	DBExtension    string `json:"database_extension"`
	ProfilesBucket string
	SessionsDB     string
	SessionsBucket string
}

// Home is the root path handler
func (rx *Remix) Home(w http.ResponseWriter, r *http.Request) {
	ss, err := rx.sess.New(r, rx.cfg.SessionName)
	if err != nil {
		// logerror
	}
	data := setSessionData(ss, rx)
	rx.rendr.HTML(w, http.StatusOK, "home", data)
}

// Register creates a ew user accounts
func (rx *Remix) Register(w http.ResponseWriter, r *http.Request) {
	ss, err := rx.sess.New(r, rx.cfg.SessionName)
	data := render.NewTemplateData()
	if err != nil {
		// Log error
	}
	if !ss.IsNew {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}
	if r.Method == "GET" {
		rx.rendr.HTML(w, http.StatusOK, "auth/register", data)
		return
	}
	if r.Method == "POST" {
		form := ComposeRegisterForm()(r)
		if !form.IsValid() {
			data.Add("errors", form.Errors())
			rx.rendr.HTML(w, http.StatusOK, "auth/register", data)
			return
		}
		user := form.GetModel().(User)
		hash, err := hashPassword(user.Pass)
		if err != nil {
			rx.rendr.HTML(w, http.StatusInternalServerError, "500", data)
			return
		}
		user.Pass = hash
		user.UUID = getUUID()
		err = CreateAccount(setDB(rx.db, rx.cfg.AccountsDB), &user, rx.cfg.AccountsBucket)
		if err != nil {
			rx.rendr.HTML(w, http.StatusInternalServerError, "500", data)
			return
		}
		flash := NewFlash()
		flash.Success("akaunti imefanikiwa kutengenezwa")
		flash.Save(ss)
		ss.Values["user"] = user.EmailAddress
		err = ss.Save(r, w)
		if err != nil {
			rx.rendr.HTML(w, http.StatusInternalServerError, "500", data)
			return
		}

		// create a new profile
		pdb := getProfileDatabase(rx.cfg.DBDir, user.UUID, rx.cfg.DBExtension)
		profile := &Profile{ID: user.UUID}
		err = CreateProfile(setDB(rx.db, pdb), profile, rx.cfg.ProfilesBucket)
		if err != nil {
			// log this
		}
		http.Redirect(w, r, rx.cfg.LoginRedirect, http.StatusFound)
		return
	}

}

// Login logges in users.
func (rx *Remix) Login(w http.ResponseWriter, r *http.Request) {
	ss, err := rx.sess.New(r, rx.cfg.SessionName)
	if err != nil {
		// log this
	}
	data := render.NewTemplateData()
	flash := NewFlash()
	if r.Method == "GET" {
		fd := flash.Get(ss)
		if fd != nil {
			data.Add("flash", fd.Data)
		}
		rx.rendr.HTML(w, http.StatusOK, "auth/login", data)
		return
	}
	if r.Method == "POST" {
		form := ComposeLoginForm()(r)
		if !form.IsValid() {
			data.Add("errors", form.Errors())
			rx.rendr.HTML(w, http.StatusOK, "auth/login", data)
			return
		}

		lform := form.GetModel().(loginForm)
		user, err := GetUser(setDB(rx.db, rx.cfg.AccountsDB), rx.cfg.AccountsBucket, lform.Email)
		if err != nil {
			data.Add("error", "email au namba ya siri sio sahihi, tafadhali jaribu tena")
			rx.rendr.HTML(w, http.StatusOK, "auth/login", data)
			return
		}
		if err = verifyPass(user.Pass, lform.Password); err != nil {
			data.Add("error", "email au namba ya siri sio sahihi, tafadhali jaribu tena")
			rx.rendr.HTML(w, http.StatusOK, "auth/login", data)
			return
		}
		ss.Values["user"] = user.EmailAddress
		err = ss.Save(r, w)
		if err != nil {
			rx.rendr.HTML(w, http.StatusInternalServerError, "500", data)
			return
		}
		http.Redirect(w, r, rx.cfg.LoginRedirect, http.StatusFound)
		return
	}
}
func setDB(db nutz.Storage, dbname string) nutz.Storage {
	d := db
	d.DBName = dbname
	return d
}

// Sets the InSession value, and and flash(which contains flash messages) to be used as
// context in templates.
func setSessionData(ss *sessions.Session, rx *Remix) render.TemplateData {
	log.SetFlags(log.Lshortfile)
	data := render.NewTemplateData()
	flash := NewFlash()
	fd := flash.Get(ss)
	if fd != nil {
		data.Add("flsh", fd.Data)
	}
	if !ss.IsNew {
		data.Add("InSession", true)
		email := ss.Values["user"].(string)
		user, err := GetUser(setDB(rx.db, rx.cfg.AccountsDB), rx.cfg.AccountsBucket, email)
		if err != nil {
			return data
		}
		pdb := getProfileDatabase(rx.cfg.DBDir, user.UUID, rx.cfg.DBExtension)
		p, err := GetProfile(setDB(rx.db, pdb), rx.cfg.ProfilesBucket, user.UUID)
		if err != nil {
			return data
		}
		data.Add("CurrentUser", user)
		data.Add("Profile", p)
		return data
	}
	return data
}
