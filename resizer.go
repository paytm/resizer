package main

import (
  "net/http"
  "log"
  "fmt"
  "os"
  "./resized"
  "github.com/codegangsta/negroni"
  "github.com/paytm/resizer/logging"
  "flag"
  "./middleware"
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

  if cfg.Server.Rate != 0 {
    fmt.Printf("Rate limiting at %d req/sec\n",cfg.Server.Rate)
    n.Use(middleware.Ratelimit(cfg.Server.Rate))
  }

  n.Use(negroni.HandlerFunc(resized.Resizer(cfg.Downstream, cfg.Upstream, cfg.Server)))
  
  n.UseHandler(mux)
  n.Run(":" + cfg.Server.Port)
}
