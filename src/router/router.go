package router

import (
  "fmt"
  "net/http"
  "os"
  "log"
  "github.com/nfnt/resize"
  "image/jpeg"
  "path"
  "strings"
  "strconv"
)

const (
  Assets = "/tmp"
  Cache = "./public"
  Base = "/images/catalog/product/"
)

func getFilePathResQuality(url string) (path string, width, height, quality int) {
  var res []string
  path = strings.TrimPrefix(url,Base)
  fields := strings.Split(path,"/")
  length := len(fields)
  path = Base + strings.Join(fields[0:3],"/") + "/" + fields[length-1]
  switch (length) {
    case 6:
      quality,_ = strconv.Atoi(fields[4])
      res = strings.Split(fields[3],"x")
    case 5:
      res = strings.Split(fields[3],"x")
    case 4:
    default:
  }
  if (res != nil) {
    width,_ = strconv.Atoi(res[0])
    height,_ = strconv.Atoi(res[1])
  }
  return
}

func Init() *http.ServeMux {
  mux := http.NewServeMux()

  mux.HandleFunc("/images/catalog/product/", func(w http.ResponseWriter, r* http.Request) {

    filePath,width,height,quality := getFilePathResQuality(r.URL.Path)

    log.Println(filePath,width,height,quality)
    log.Println("You requested for ", Assets + filePath);

    file, err := os.Open(Assets + filePath);
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

    m := resize.Thumbnail(uint(width), uint(height), img, resize.Lanczos3)
    q := jpeg.Options{ Quality: quality }
    jpeg.Encode(w,m, &q)

    // cache the result as well, on disk
    cachePath := Cache + r.URL.Path
    err = os.MkdirAll(path.Dir(cachePath),os.ModeDir | 0777)
    if err == nil {
      out, _ := os.Create(cachePath)
      jpeg.Encode(out,m,&q)
      out.Close()
      log.Println("cached into " + cachePath);
    } else {
      log.Println("cache fail ",err.Error())
    }
  })

  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  return mux
}
