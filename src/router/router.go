package router

import (
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
  Base = "/images/catalog/product/"
)

const (
  PathComponentsMax = 3
  QualityIndex = 4
  ResolutionIndex = 3
)

type HandlerFunc func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)

func getFilePathResQuality(url string) (path string, width, height, quality int) {
  var res []string
  path = strings.TrimPrefix(url,Base)
  fields := strings.Split(path,"/")
  length := len(fields)
  path = Base + strings.Join(fields[:PathComponentsMax],"/") + "/" + fields[length-1]

  // defaults
  quality = 70
  width = 0
  height = 0

  switch (length) {
    case 6:
      quality,_ = strconv.Atoi(fields[QualityIndex])
      res = strings.Split(fields[ResolutionIndex],"x")
    case 5:
      res = strings.Split(fields[ResolutionIndex],"x")
    case 4:
    default:
  }

  if (res != nil) {
    width,_ = strconv.Atoi(res[0])
    height,_ = strconv.Atoi(res[1])
  }
  return
}

/*
type Upstream interface {
  Serve(w http.ResponseWriter, r *http.Request, path string) 
  Get() 
}
*/

func serveWithoutResize(w http.ResponseWriter,r *http.Request,path string) {
  http.ServeFile(w,r,path)
}

func Resizer(cacheDir string,upstream string) (HandlerFunc) {

  return func(w http.ResponseWriter, r* http.Request, next http.HandlerFunc) {

    if (strings.HasPrefix(r.URL.Path,"/images/catalog/product/") == false) {
      log.Println("skipping ",r.URL.Path)
      next(w,r);
      return
    }

    filePath,width,height,quality := getFilePathResQuality(r.URL.Path)

    if (width == 0 && height == 0) {
      log.Println("skipping resize ",r.URL.Path)
      serveWithoutResize(w,r,Assets + filePath)
      return
    }

/*
    file, err := os.Open(Assets + filePath);
    defer file.Close()

    if err != nil {
      http.Error(w, err.Error(), http.StatusNotFound)
      return
    }
*/

    resp,err := http.Get(upstream + filePath)
    if (err != nil || resp.StatusCode != 200) {
      log.Println("upstream error with ",r.URL.Path)
      http.Error(w, "File not found", http.StatusNotFound)
      return
    }
    file := resp.Body
    defer file.Close()

    img, err := jpeg.Decode(file)
    if err != nil {
      log.Println("Failed to decode jpeg ")
      http.Error(w, err.Error(), http.StatusNotFound)
      return
    }

    m := resize.Thumbnail(uint(width), uint(height), img, resize.Lanczos3)
    q := jpeg.Options{ Quality: quality }
    jpeg.Encode(w,m, &q)

    // cache the result as well, on disk
    if (cacheDir != "") {
      cachePath := cacheDir + r.URL.Path
      err = os.MkdirAll(path.Dir(cachePath),os.ModeDir | 0777)
      if err == nil {
        out, _ := os.Create(cachePath)
        jpeg.Encode(out,m,&q)
        out.Close()
        log.Println("cached into " + cachePath);
      } else {
        log.Println("cache fail ",err.Error())
      }
    }
  }
}
