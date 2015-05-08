package aurora

import (
	"net/http"

	"github.com/gernest/nutz"

	"github.com/gernest/render"
)

// Remix all the fun is here
type Remix struct {
	sess       *Session
	rendr      *render.Render
	accoundtDB nutz.Storage
	cfg        *RemixConfig
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

	// The path to point to when login is success
	LoginRedirect string `json:"login_redirect"`
}

// Home is the root path handler
func (rx *Remix) Home(w http.ResponseWriter, r *http.Request) {
	ss, err := rx.sess.New(r, rx.cfg.SessionName)
	if err != nil {
		// logerror
	}
	data := setSessionData(ss)
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
		err = CreateAccount(rx.accoundtDB, &user, rx.cfg.AccountsBucket)
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
		user, err := GetUser(rx.accoundtDB, rx.cfg.AccountsBucket, lform.Email)
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
