package aurora

import (
	"fmt"

	"github.com/gernest/render"

	"github.com/gernest/mrs"
	"github.com/gernest/warlock"
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
	fmt.Println("aurora")
}
