package aurora

import (
	"net/http"

	"github.com/gernest/nutz"

	"github.com/gernest/render"
)

type Remix struct {
	sess       *Session
	rendr      *render.Render
	accoundtDB nutz.Storage
	cfg        *RemixConfig
}
type RemixConfig struct {
	AppName        string `json:"name"`
	AppUrl         string `json:"url"`
	CdnMode        bool   `json:"cdn_mode"`
	RunMode        string `json:"run_mode"`
	AppTitle       string `json:"title"`
	AppDescription string `json:"description"`
	SessionName    string `json:"session_name"`
	AccountsBucket string `json:"accounts_bucket"`
}

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
		flash.Add(ss)
		ss.Values["user"] = user.Email()
		err = ss.Save(r, w)
		if err != nil {
			rx.rendr.HTML(w, http.StatusInternalServerError, "500", data)
			return
		}
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

}
