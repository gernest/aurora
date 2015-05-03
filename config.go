package aurora

import (
	"github.com/fatih/structs"
	"github.com/gernest/render"
)

type Config struct {
	Appname        string `json:"name"`
	AppUrl         string `json:"url"`
	CdnMode        bool   `json:"cdn_mode"`
	RunMode        string `json:"run_mode"`
	AppTitle       string `json:"title"`
	AppDescription string `json:"description"`
}

func (c *Config) TemplData() render.TemplateData {
	d := render.NewTemplateData()
	d.Merge(structs.Map(c))
	return d
}
