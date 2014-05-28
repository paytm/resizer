package router

import (
  "fmt"
  "net/http"
  "os"
  "log"
  "github.com/nfnt/resize"
  "image/jpeg"
  "path"
)

const (
  assets = "/tmp"
  cache = "./public"
)

func Init() *http.ServeMux {
  mux := http.NewServeMux()

  mux.HandleFunc("/images/catalog/product/", func(w http.ResponseWriter, r* http.Request) {
    filePath :=  assets + r.URL.Path
    log.Println("You requested for ", filePath);

    file, err := os.Open(filePath);
    defer file.Close()

    if err != nil {
      http.Error(w, err.Error(), http.StatusNotFound)
      return
    }

    img, err := jpeg.Decode(file)
    if err != nil {
      http.Error(w, err.Error(), http.StatusNotFound)
      return
    }

    m := resize.Thumbnail(100, 100, img, resize.Lanczos3)
    jpeg.Encode(w,m, &jpeg.Options{ Quality: 70 })

    // cache the result as well, on disk
    cachePath := cache + r.URL.Path
    err = os.MkdirAll(path.Dir(cachePath),os.ModeDir)
    if err != nil {
      out, _ := os.Create(cachePath)
      jpeg.Encode(out,m,nil)
      out.Close()
      log.Println("cached into " + cachePath);
    }
  })

  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  return mux
}
