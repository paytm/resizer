package main

import (
  "net/http"
  "log"
  "os"
  "github.com/qzaidi/resizer/resized"
  "github.com/codegangsta/negroni"
  "code.google.com/p/gcfg"
)

type Config struct {
  Upstream struct {
    URI string
  }
  Downstream struct {
    URI string
    MaxThreads int
  }
  Server struct {
   Port string
  }
}

func readConfig(cfg *Config,path string) bool {
  err := gcfg.ReadFileInto(cfg,path + "/resizer.ini")
  if err == nil {
    log.Println("read config from ",path)
    return true
  }
  return false
}

func main() {

  var cfg Config
  ok := readConfig(&cfg, ".") || readConfig(&cfg,"/etc")
  if (!ok) {
    log.Println("failed to read resizer.ini from CWD or /etc")
    os.Exit(1)
  }

  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    http.Error(w, "File not found", http.StatusNotFound)
  })

  n := negroni.Classic()
  n.Use(negroni.HandlerFunc(resized.Resizer(cfg.Downstream.URI, cfg.Downstream.MaxThreads, cfg.Upstream.URI)))
  n.UseHandler(mux)
  n.Run(":" + cfg.Server.Port)
}
