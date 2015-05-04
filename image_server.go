package aurora

import (
	"bytes"
	"net/http"

	"github.com/gernest/mrs"
	"github.com/gorilla/mux"
)

type ImageServer struct {
	pm *mrs.PhotoManager
}

func NewImageServer(cfg *mrs.Config) *ImageServer {
	return &ImageServer{mrs.NewPhotoManager(cfg.DB, cfg.MetaBucket, cfg.DataBucket)}
}

func (i *ImageServer) Show(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid := vars["id"]
	p, d, err := i.pm.GetPhoto(pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	picName := p.ID + "." + p.Type
	http.ServeContent(w, r, picName, p.UpdatedAt, bytes.NewReader(d))
	return
}
