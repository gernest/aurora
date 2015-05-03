package aurora

import (
	"log"
	"net/http"

	"github.com/gernest/render"

	"github.com/gernest/mrs"
	"github.com/gernest/warlock"
	"github.com/gorilla/mux"
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

func (a *Aurora) Run() {
	m := mux.NewRouter()
	m.HandleFunc("/", a.Base.Base)
	port := ":8080"
	log.Println("listening at localhost port ", port)
	log.Fatal(http.ListenAndServe(port, m))
}
