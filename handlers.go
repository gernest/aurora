package aurora

import (
	"net/http"

	"github.com/gernest/render"
	"github.com/gorilla/context"
)

type Handlers struct {
	rendr *render.Render
	cfg   *Config
}

func NewHandlers(cfg *Config, r *render.Render) *Handlers {
	return &Handlers{r, cfg}
}
func (h *Handlers) Base(w http.ResponseWriter, r *http.Request) {
	d := h.setSessData(r)
	h.rendr.HTML(w, http.StatusOK, "home", d)
}
func (h *Handlers) setSessData(r *http.Request) render.TemplateData {
	u := context.Get(r, "user")
	if u != nil {
		td := render.NewTemplateData()
		td.Add("user", u)
		td.Add("InSession", true)
		return td
	}
	return nil
}
