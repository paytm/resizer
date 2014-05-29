package router

import (
  "net/http"
  "net/url"
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
  Base = "/images/catalog/"
)

const (
  PathComponentsProductMax = 4
  PathComponentsCategoryMax = 2
  QualityIndex = 5
  ResolutionIndex = 4
)

type HandlerFunc func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)

func getFilePathResQuality(url string) (path string, width, height, quality int) {
  var res []string
  var resq []string
  path = strings.TrimPrefix(url,Base)
  fields := strings.Split(path,"/")
  length := len(fields)

  if fields[0] == "category"  {
    path = Base +  strings.Join(fields[:PathComponentsCategoryMax],"/") + "/" + fields[length-1];
    resq = fields[PathComponentsCategoryMax:length-1]
  } else {
    path = Base + strings.Join(fields[:PathComponentsProductMax],"/") + "/" + fields[length-1]
    resq = fields[PathComponentsProductMax:length-1]
  }

  log.Println(resq);
  log.Println(path);

  // defaults
  quality = 70
  width = 0
  height = 0

  switch (len(resq)) {
    case 2:
      quality,_ = strconv.Atoi(resq[1])
      res = strings.Split(resq[0],"x")
    case 1:
      res = strings.Split(resq[0],"x")
    default:
  }

  log.Println(res)

  if (res != nil) {
    width,_ = strconv.Atoi(res[0])
    height,_ = strconv.Atoi(res[1])
  }
  return
}

func Resizer(cacheDir string,ups string) (HandlerFunc) {

  var server Upstream

  url,err := url.Parse(ups)
  if err != nil {
    log.Panic("Bad URL scheme")
  }

  switch url.Scheme {
    case "http":
      server = &HTTPUpstream{ upstreamURI: ups}
      log.Println("Serving using " + ups)
    case "file":
      server = &FileUpstream{ upstreamURI: url.Path }
      log.Println("Serving using " + url.Path)
    default:
      log.Panic("Unsupported url scheme " + url.Scheme)
  }

  return func(w http.ResponseWriter, r* http.Request, next http.HandlerFunc) {

    if (strings.HasPrefix(r.URL.Path,"/images/catalog/") == false) {
      log.Println("skipping ",r.URL.Path)
      next(w,r);
      return
    }

    filePath,width,height,quality := getFilePathResQuality(r.URL.Path)

    if (width == 0 && height == 0) {
      log.Println("skipping resize ",r.URL.Path)
      server.ServeOriginal(w,r,filePath)
      return
    }


    file,err := server.Get(w,r,filePath)
    if (err != nil) {
      log.Println("upstream error with ",r.URL.Path)
      http.Error(w, "File not found", http.StatusNotFound)
      return
    } else {
      defer file.Close()
    }

    img, err := jpeg.Decode(file)
    if err != nil {
      log.Println("Failed to decode jpeg ")
      http.Error(w, err.Error(), http.StatusNotFound)
      return
    }

    m := resize.Resize(uint(width), uint(height), img, resize.NearestNeighbor)
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
