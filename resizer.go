package main

import (
  "net/http"
  "log"
  "resizer"
  "github.com/codegangsta/negroni"
  "code.google.com/p/gcfg"
)

type Config struct {
  Upstream struct {
    URI string
  }
  Downstream struct {
    URI string
  }
  Server struct {
   Port string
  }
}

func main() {

  var cfg Config
  err := gcfg.ReadFileInto(&cfg,"resizer.ini")
  if (err != nil) {
    log.Println("failed to read config ",err.Error())
    cfg.Server.Port = "3000"
    cfg.Upstream.URI = "file:///tmp"
  }

  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    http.Error(w, "File not found", http.StatusNotFound)
  })

  n := negroni.Classic()
  n.Use(negroni.HandlerFunc(resizer.Resizer(cfg.Downstream.URI,cfg.Upstream.URI)))
  n.UseHandler(mux)
  n.Run(":" + cfg.Server.Port)
}
