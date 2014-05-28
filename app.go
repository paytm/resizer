package main

import (
  "github.com/codegangsta/negroni"
  "router"
)

const (
  CacheDir = "./public"
)

func main() {
  mux := router.Init(CacheDir)

  n := negroni.Classic()
  n.UseHandler(mux)
  n.Run(":3000")
}
