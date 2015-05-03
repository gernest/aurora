package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"path/filepath"

	"github.com/gernest/aurora"
	"github.com/gernest/mrs"
	"github.com/gernest/warlock"
)

func loadConfigs(conf string) (*warlock.Config, *mrs.Config, *aurora.Config) {
	w := &warlock.Config{}
	m := &mrs.Config{}
	a := &aurora.Config{}
	dw, err := ioutil.ReadFile(filepath.Join(conf, "app/warlock.json"))
	if err != nil {
		panic(err)
	}
	dm, err := ioutil.ReadFile(filepath.Join(conf, "app/mrs.json"))
	if err != nil {
		panic(err)
	}
	da, err := ioutil.ReadFile(filepath.Join(conf, "app/app.json"))
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(dw, w)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(dm, m)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(da, a)
	if err != nil {
		panic(a)
	}
	return w, m, a
}
func main() {
	conf := flag.String("c", "config", "specifies the directory where config files are")
	flag.Parse()
	w, m, a := loadConfigs(*conf)
	app := aurora.New(a, w, m)
	app.Run()
}
