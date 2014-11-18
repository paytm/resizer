package main

import (
  "net/http"
  "log"
  "fmt"
  "os"
  // "github.com/paytm/resizer/resized"
  "./resized"
  "github.com/codegangsta/negroni"
  "github.com/paytm/resizer/logging"
  "flag"
  "./wrapper"
)

func main() {

  var cfg resized.Config
  ok := resized.ReadConfig(&cfg, ".") || resized.ReadConfig(&cfg,"/etc")
  if (!ok) {
    log.Println("failed to read resizer.ini from CWD or /etc ")
    os.Exit(1)
  }

  v := flag.Bool("version",false,"prints resizer version")
  logging.Init() // this sets -e & -l flags
  flag.Parse()

  if *v {
    fmt.Println(ResizerVersion())
    os.Exit(0)
  }

  logging.LogInit()
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    http.Error(w, "File not found", http.StatusNotFound)
  })

  n := negroni.Classic()
  n.Use(negroni.HandlerFunc(wrapper.RateLimit(resized.Resizer(cfg.Downstream, cfg.Upstream, cfg.Server),cfg.Server.Rate)))
  n.UseHandler(mux)
  n.Run(":" + cfg.Server.Port)
}
