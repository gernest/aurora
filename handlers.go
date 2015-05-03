package aurora

import (
	"net/http"

	"github.com/gernest/render"
)

type Handlers struct {
	rendr *render.Render
	cfg   *Config
}

func NewHandlers(cfg *Config, r *render.Render) *Handlers {
	return &Handlers{r, cfg}
}
func (h *Handlers) Base(w http.ResponseWriter, r *http.Request) {
	h.rendr.HTML(w, http.StatusOK, "home", nil)
}
