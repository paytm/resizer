package main

import (
  "github.com/codegangsta/negroni"
  "net/http"
  "router"
  "code.google.com/p/gcfg"
)

const (
  CacheDir = "./public"
  AssetServer = "http://assets.paytm.com"
)

type Config struct {
  Upstream struct {
    URI string
  }
  Server struct {
   Port string
   CacheDir string
  }
}

func main() {

  var cfg Config
  err := gcfg.ReadFileInto(&cfg,"resizer.ini")
  if (err != nil) {
    cfg.Server.Port = "3000"
    cfg.Upstream.URI = "file:///tmp"
    cfg.Server.CacheDir = ""
  }

  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    http.Error(w, "File not found", http.StatusNotFound)
  })

  n := negroni.Classic()
  n.Use(negroni.HandlerFunc(router.Resizer(cfg.Server.CacheDir,cfg.Upstream.URI)))
  n.UseHandler(mux)
  n.Run(":" + cfg.Server.Port)
}
