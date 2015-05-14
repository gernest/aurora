package aurora

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	"github.com/gernest/nutz"
	"github.com/gernest/render"
	"github.com/gorilla/sessions"
)

var (
	errNotFound       error = errors.New("samahani kitu ulichoulizia hatujakipata")
	errInternalServer error = errors.New("du! naona imezingua, jaribu tena badae")
	errForbidden      error = errors.New("du! hauna ruhususa ya kufika kwenye hii kurasa")
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

	// path to the directory where databases will be stored
	DBDir string `json:"database_dir"`

	AccountsBucket string `json:"accounts_bucket"`
	AccountsDB     string `json:"accounts_database"`
	DBExtension    string `json:"database_extension"`
	ProfilesBucket string `json:"profiles_bucket"`

	SessionName    string `json:"sessions_name"`
	SessionsDB     string `json:"sessions_database"`
	SessionsBucket string `json:"sessions_bucket"`
	SessMaxAge     int    `json:"sessions_max_age"`
	SessionPath    string `json:"session_path"`

	// The path to point to when login is success
	LoginRedirect string `json:"login_redirect"`

	ProfilePicField string `json:"profile_pic_field"`
	PhotosField     string `json:"photos_field"`

	MessagesBucket string `json:"messages_bucket"`

	TemplatesExtensions []string `json:"templates_extensions"`
	TemplatesDir        string   `json:"templates_dir"`
	DevMode             bool     `json:"dev_mode"`
}

type jsonUploads struct {
	Error      string   `json:"errors"`
	ProfilePic *photo   `json:"profile_photo"`
	Photos     []*photo `json:"photos"`
}
type jsonErr struct {
	Text string `json:"test"`
}

func NewRemix(cfg *RemixConfig) *Remix {
	secret := []byte("my-top-secret")
	rOpts := render.Options{
		Directory:     cfg.TemplatesDir,
		Extensions:    cfg.TemplatesExtensions,
		IsDevelopment: cfg.DevMode,
		DefaultData:   setConfigData(cfg),
	}
	sOpts := &sessions.Options{
		MaxAge: cfg.SessMaxAge,
		Path:   cfg.SessionPath,
	}
	db := nutz.NewStorage(cfg.SessionsDB, 0600, nil)
	store := NewSessStore(db, cfg.SessionsBucket, 10, sOpts, secret)
	rx := &Remix{
		db:    db,
		sess:  store,
		rendr: render.New(rOpts),
		cfg:   cfg,
	}
	return rx
}

// Home is the root path handler
func (rx *Remix) Home(w http.ResponseWriter, r *http.Request) {
	data := rx.setSessionData(r)
	rx.rendr.HTML(w, http.StatusOK, "home", data)
}

// Register creates a ew user accounts
func (rx *Remix) Register(w http.ResponseWriter, r *http.Request) {
	var (
		ss   *sessions.Session
		ok   bool
		data render.TemplateData = render.NewTemplateData()
	)
	if ss, ok = rx.isInSession(r); ok {
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
		ss.Values["isAuthorized"] = true
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

// Login creates new session for a user
func (rx *Remix) Login(w http.ResponseWriter, r *http.Request) {
	var (
		ss    *sessions.Session
		ok    bool
		data  render.TemplateData = render.NewTemplateData()
		flash *Flash              = NewFlash()
	)

	if ss, ok = rx.isInSession(r); ok {
		http.Redirect(w, r, rx.cfg.LoginRedirect, http.StatusFound)
		return
	}
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
		ss, err = rx.sess.New(r, rx.cfg.SessionName)
		if err != nil {
			//log this
		}
		ss.Values["user"] = user.EmailAddress
		ss.Values["isAuthorized"] = true
		err = ss.Save(r, w)
		if err != nil {
			rx.rendr.HTML(w, http.StatusInternalServerError, "500", data)
			return
		}
		http.Redirect(w, r, rx.cfg.LoginRedirect, http.StatusFound)
		return
	}
}

// ServeImages serves images uploaded by users
func (rx *Remix) ServeImages(w http.ResponseWriter, r *http.Request) {
	var (
		vars        url.Values = r.URL.Query()
		imageID     string     = vars.Get("iid")
		profileID   string     = vars.Get("pid")
		pic         *photo     = &photo{}
		photoBucket string     = "photos"
		metaBucket  string     = "meta"
		dataBucket  string     = "data"
	)

	pdb := getProfileDatabase(rx.cfg.DBDir, profileID, rx.cfg.DBExtension)
	db := setDB(rx.db, pdb)

	err := getAndUnmarshall(db, photoBucket, imageID, pic, metaBucket)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	raw := db.Get(photoBucket, imageID, dataBucket)
	if raw.Error != nil {
		http.NotFound(w, r)
		return
	}
	picName := fmt.Sprintf("%s.%s", pic.ID, pic.Type)
	http.ServeContent(w, r, picName, pic.UpdatedAt, bytes.NewReader(raw.Data))
}

// Uploads uploads files to the uploader's database
func (rx *Remix) Uploads(w http.ResponseWriter, r *http.Request) {
	var (
		ss   *sessions.Session
		ok   bool
		rst  []*photo
		errs listErr
	)
	if ss, ok = rx.isInSession(r); !ok {
		jr := &jsonUploads{Error: "not authorized"}
		rx.rendr.JSON(w, http.StatusForbidden, jr)
		return
	}
	if r.Method == "POST" {
		_, profile, err := rx.getCurrentUserAndProfile(ss)
		if err != nil {
			jr := &jsonUploads{Error: err.Error()}
			rx.rendr.JSON(w, http.StatusInternalServerError, jr)
			return
		}

		pdbStr := getProfileDatabase(rx.cfg.DBDir, profile.ID, rx.cfg.DBExtension)
		pdb := setDB(rx.db, pdbStr)

		f, serr := GetFileUpload(r, rx.cfg.ProfilePicField)
		if serr == nil {
			pic, err := SaveUploadFile(pdb, f, profile)
			if err != nil {
				jr := &jsonUploads{Error: err.Error()}
				rx.rendr.JSON(w, http.StatusInternalServerError, jr)
				return
			}
			profile.Picture = pic
			err = UpdateProfile(pdb, profile, rx.cfg.ProfilesBucket)
			if err != nil {
				jr := &jsonUploads{Error: err.Error()}
				rx.rendr.JSON(w, http.StatusInternalServerError, jr)
				return
			}
			rx.rendr.JSON(w, http.StatusOK, pic)
			return
		}

		files, ferr := GetMultipleFileUpload(r, rx.cfg.PhotosField)
		if ferr != nil && len(files) > 0 || err == nil && len(files) > 0 {
			for _, v := range files {
				pic, err := SaveUploadFile(pdb, v, profile)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				rst = append(rst, pic)
			}
			errs = append(errs, ferr)
			if len(rst) == 0 && len(errs) > 0 {
				jr := &jsonUploads{Error: err.Error()}
				rx.rendr.JSON(w, http.StatusInternalServerError, jr)
				return
			}
			profile.Photos = rst
			err = UpdateProfile(pdb, profile, rx.cfg.ProfilesBucket)
			if err != nil {
				jr := &jsonUploads{Error: err.Error()}
				rx.rendr.JSON(w, http.StatusInternalServerError, jr)
				return
			}
			if serr != nil {
				errs = append(errs, serr)
			}
			jr := &jsonUploads{Error: errs.Error(), Photos: rst}
			rx.rendr.JSON(w, http.StatusOK, jr)
			return
		}
		if serr != nil {
			jr := &jsonUploads{Error: serr.Error()}
			rx.rendr.JSON(w, http.StatusInternalServerError, jr)
			return
		}
	}
}

// Logout deletes current session
func (rx *Remix) Logout(w http.ResponseWriter, r *http.Request) {
	if ss, ok := rx.isInSession(r); ok && ss != nil {
		err := rx.sess.Delete(r, w, ss)
		if err != nil {
			// log this
		}
		http.Redirect(w, r, rx.cfg.LoginRedirect, http.StatusFound)
		return
	}
}

// Profile viewing and updating profile
func (rx *Remix) Profile(w http.ResponseWriter, r *http.Request) {
	var (
		vars   url.Values          = r.URL.Query()
		data   render.TemplateData = rx.setSessionData(r)
		id     string              = vars.Get("id")
		view   string              = vars.Get("view")
		all    string              = vars.Get("all")
		update string              = vars.Get("u")
		flash  *Flash
		ss     *sessions.Session
		ok     bool
	)

	if r.Method == "GET" {
		pdb := getProfileDatabase(rx.cfg.DBDir, id, rx.cfg.DBExtension)
		if id != "" && view == "true" && all != "true" {
			if rx.isAjax(r) {
				p, err := GetProfile(setDB(rx.db, pdb), rx.cfg.ProfilesBucket, id)
				if err != nil {
					// TODO: log this err
					rx.rendr.JSON(w, http.StatusNotFound, &jsonErr{errNotFound.Error()})
					return
				}
				rx.rendr.JSON(w, http.StatusOK, p)
			}
			p, err := GetProfile(setDB(rx.db, pdb), rx.cfg.ProfilesBucket, id)
			if err != nil {
				data.Add("error", errNotFound)
				rx.rendr.HTML(w, http.StatusNotFound, "404", data)
				return
			}
			data.Add("profile", p)
			rx.rendr.HTML(w, http.StatusOK, "profile/home", data)
			return
		}
		if all == "true" && view == "true" {
			p, err := rx.getAllProfiles()
			if rx.isAjax(r) {
				if err != nil {
					// TODO: log this err
					rx.rendr.JSON(w, http.StatusNotFound, &jsonErr{errNotFound.Error()})
					return
				}
				if p != nil {
					rx.rendr.JSON(w, http.StatusOK, p)
					return
				}

			}
			if err != nil {
				data.Add("error", errNotFound)
				rx.rendr.HTML(w, http.StatusNotFound, "404", data)
				return
			}
			data.Add("profiles", p)
			rx.rendr.HTML(w, http.StatusOK, "profile/home", data)
			return
		}
	}
	if r.Method == "POST" {
		if update == "true" {
			//form := ComposeProfileForm()(r)
			if ss, ok = rx.isInSession(r); ok {
				_, p, err := rx.getCurrentUserAndProfile(ss)
				if err != nil {
					if rx.isAjax(r) {
						rx.rendr.JSON(w, http.StatusInternalServerError, &jsonErr{errInternalServer.Error()})
						return
					}
					data.Add("error", errInternalServer.Error())
					rx.rendr.HTML(w, http.StatusInternalServerError, "500", data)
					return

				}
				if p.ID != id {
					if rx.isAjax(r) {
						rx.rendr.JSON(w, http.StatusForbidden, &jsonErr{errForbidden.Error()})
						return
					}
					data.Add("error", errForbidden.Error())
					rx.rendr.HTML(w, http.StatusInternalServerError, "403", data)
					return
				}

			}
			if rx.isAjax(r) {
				rx.rendr.JSON(w, http.StatusForbidden, &jsonErr{errForbidden.Error()})
				return
			}
			flash = NewFlash()
			flash.Error("unatakiwa uingie kwanza kabla ya kupata ruhusa ya kutumia hii kurasa")
			flash.Save(ss)
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}
	}
}

func (rx *Remix) getAllProfiles() ([]*Profile, error) {
	var rst []*Profile
	usrs, err := GetAllUsers(setDB(rx.db, rx.cfg.AccountsDB), rx.cfg.AccountsBucket)
	if err != nil {
		return nil, err
	}
	for _, v := range usrs {
		pdb := getProfileDatabase(rx.cfg.DBDir, v, rx.cfg.DBExtension)
		p, err := GetProfile(setDB(rx.db, pdb), rx.cfg.ProfilesBucket, v)
		if err != nil {
			// log this
		}
		if p != nil {
			rst = append(rst, p)
		}
	}
	if len(rst) == 0 {
		return nil, errNotFound
	}
	return rst, nil
}

// Routes returs a mux of all registered routes
func (rx *Remix) Routes() *mux.Router {
	h := mux.NewRouter()
	h.HandleFunc("/", rx.Home)
	h.HandleFunc("/auth/register", rx.Register).Methods("GET", "POST")
	h.HandleFunc("/auth/login", rx.Login).Methods("GET", "POST")
	h.HandleFunc("/auth/logout", rx.Logout)
	h.HandleFunc("/imgs", rx.ServeImages).Methods("GET")
	h.HandleFunc("/uploads", rx.Uploads)
	h.HandleFunc("/profile", rx.Profile)
	return h
}

func (rx *Remix) isInSession(r *http.Request) (*sessions.Session, bool) {
	var (
		ss  *sessions.Session
		err error
	)
	if ss, err = rx.sess.Get(r, rx.cfg.SessionName); err == nil {
		if v, ok := ss.Values["isAuthorized"]; ok && v == true {
			return ss, true
		}
	}
	return ss, false
}
func (rx *Remix) setSessionData(r *http.Request) render.TemplateData {
	var (
		data  render.TemplateData = render.NewTemplateData()
		flash *Flash              = NewFlash()
	)
	if ss, ok := rx.isInSession(r); ok {
		fd := flash.Get(ss)
		if fd != nil {
			data.Add("flash", fd.Data)
		}
		data.Add("InSession", true)
		user, p, err := rx.getCurrentUserAndProfile(ss)
		if err != nil {
			return data
		}
		data.Add("CurrentUser", user)
		data.Add("Profile", p)
		return data
	}
	return data
}
func (rx *Remix) getCurrentUserAndProfile(ss *sessions.Session) (*User, *Profile, error) {
	if e, ok := ss.Values["user"]; ok {
		email := e.(string)
		user, err := GetUser(setDB(rx.db, rx.cfg.AccountsDB), rx.cfg.AccountsBucket, email)
		if err != nil {
			return nil, nil, err
		}
		pdb := getProfileDatabase(rx.cfg.DBDir, user.UUID, rx.cfg.DBExtension)
		p, err := GetProfile(setDB(rx.db, pdb), rx.cfg.ProfilesBucket, user.UUID)
		if err != nil {
			return nil, nil, err
		}
		return user, p, nil
	}
	return nil, nil, errors.New("aurora: session values not set")
}

// switches databases
func setDB(db nutz.Storage, dbname string) nutz.Storage {
	d := db
	d.DBName = dbname
	return d
}

// Sets basic configuration values which has use to the templates
func setConfigData(c *RemixConfig) render.TemplateData {
	data := render.NewTemplateData()
	data.Add("AppName", c.AppName)
	data.Add("AppTitle", c.AppTitle)
	data.Add("AppDescription", c.AppDescription)
	data.Add("CdnMode", c.CdnMode)
	data.Add("AppURL", c.AppURL)
	data.Add("RunMode", c.RunMode)
	return data

}

func (rx *Remix) isAjax(r *http.Request) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}
