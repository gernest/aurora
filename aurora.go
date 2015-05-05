package aurora

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gernest/mrs"
	"github.com/gernest/render"
	"github.com/gernest/warlock"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

type Aurora struct {
	Auth    *warlock.Handlers
	Profile *mrs.Handlers
	Base    *Handlers
	Photo   *ImageServer
}

func New(aCfg *Config, wCfg *warlock.Config, mCfg *mrs.Config) *Aurora {
	tmplData := render.NewTemplateData()
	tmplData.Merge(aCfg.TemplData())
	renderConfig := render.Options{
		Directory:     "templates",
		Extensions:    []string{".tmpl", ".html", ".tpl"},
		IsDevelopment: true,
		DefaultData:   tmplData,
	}
	rendr := render.New(renderConfig)
	w := warlock.YoungWarlock(rendr, wCfg)
	m := mrs.NewHandlers(mCfg, &renderConfig, rendr)
	h := NewHandlers(aCfg, rendr)
	p := NewImageServer(mCfg)
	return &Aurora{w, m, h, p}
}
func (a *Aurora) Routes() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc("/", a.Base.Base).Methods("GET")

	// auth
	m.HandleFunc("/auth/register", a.Auth.Register).Methods("GET", "POST")
	m.HandleFunc("/auth/login", a.Auth.Login).Methods("GET", "POST")
	m.HandleFunc("/auth/logout", a.Auth.Logout).Methods("GET", "POST")

	// profile
	pid := "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"
	m.HandleFunc(fmt.Sprintf("/profile/{id:%s}", pid), a.Profile.MeOnly(a.Profile.Home)).Methods("GET", "POST")
	m.HandleFunc(fmt.Sprintf("/profile/pic/{id:%s}", pid), a.Profile.ProfilePic).Methods("POST")
	m.HandleFunc(fmt.Sprintf("/profile/uploads/{id:%s}", pid), a.Profile.FileUploads).Methods("POST")
	m.HandleFunc(fmt.Sprintf("/profile/view/{id:%s}", pid), a.Profile.View).Methods("POST")

	// photo
	m.HandleFunc(fmt.Sprintf("/imgs/{id:%s}", pid), a.Photo.Show)
	return m
}

func (a *Aurora) Run() {
	port := ":8080"
	log.Println("listening at localhost port ", port)
	stack := alice.New(a.Auth.SessionMiddleware)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
	http.Handle("/", stack.Then(a.Routes()))
	log.Fatal(http.ListenAndServe(port, nil))

}
