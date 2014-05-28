package main

import (
  "github.com/codegangsta/negroni"
  "router"
  "net/http"
)

func main() {
  mux := router.Init()

  n := negroni.Classic()
  n.Use(negroni.NewStatic(http.Dir("./public")))
  n.UseHandler(mux)
  n.Run(":3000")
}
