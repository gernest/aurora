package aurora

import (
	"log"
	"net/http"

	"github.com/gernest/render"

	"github.com/gernest/mrs"
	"github.com/gernest/warlock"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

type Aurora struct {
	Auth    *warlock.Handlers
	Profile *mrs.Handlers
	Base    *Handlers
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
	return &Aurora{w, m, h}
}
func (a *Aurora) Routes() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc("/", a.Base.Base).Methods("GET")
	m.HandleFunc("/auth/register", a.Auth.Register).Methods("GET", "POST")
	m.HandleFunc("/auth/login", a.Auth.Login).Methods("GET", "POST")
	m.HandleFunc("/auth/logout", a.Auth.Logout).Methods("GET", "POST")
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
