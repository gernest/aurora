package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gernest/aurora"
)

func main() {
	d, err := ioutil.ReadFile("config/app.json")
	if err != nil {
		panic(err)
	}
	cfg := &aurora.RemixConfig{}
	err = json.Unmarshal(d, cfg)
	if err != nil {
		panic(err)
	}
	rx := aurora.NewRemix(cfg)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
	http.Handle("/", rx.Routes())
	log.Println("starting server ar port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
