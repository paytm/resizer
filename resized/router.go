/*
 Package resized implements a negroni middleware for on the fly resizing.
 It uses magickwand to resize, and supports a file/http origin to fetch the
 originals from. Resized images can be optionally saved to a file/s3 downstream.
*/
package resized

import (
  "net/http"
  "net/url"
  "github.com/gographics/imagick/imagick"
  "log"
  "strings"
  "strconv"
  "mime"
  "errors"
  "io/ioutil"
  "regexp"
  "time"
  "os"
  "fmt"
)

// These constants define the structure of a resize url
const (
  Base = "/images/catalog/"
  PathComponentsProductMax = 4
  PathComponentsCategoryMax = 2
  QualityIndex = 5
  ResolutionIndex = 4
)

type HandlerFunc func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)

var sizeRegex *regexp.Regexp

func validateExtensions(ext string) {
  for _,typ := range(strings.Split(ext," ")) {
    if ( mime.TypeByExtension("." + typ) == "") {
      log.Println("No Mime type configured for " + typ)
      os.Exit(1)
    }
  }
}

func getFilePathResQuality(url string) (err error,path string, width, height, quality int) {
  var res []string
  var resq []string
  fields := strings.Split(strings.TrimPrefix(url,Base),"/")
  length := len(fields)

  // defaults
  quality = 70
  width = 0
  height = 0

  if fields[0] == "category"  {
    if (length >= PathComponentsCategoryMax) {
      path = Base +  strings.Join(fields[:PathComponentsCategoryMax],"/") + "/" + fields[length-1];
      resq = fields[PathComponentsCategoryMax:length-1]
    }
  } else if fields[0] == "product" {
    if (length > PathComponentsProductMax) {
      path = Base + strings.Join(fields[:PathComponentsProductMax],"/") + "/" + fields[length-1]
      resq = fields[PathComponentsProductMax:length-1]
    }
  } else if matches := sizeRegex.FindString(url); matches != "" {
     resq = []string { strings.TrimPrefix(matches,"/") }
     path = strings.Join(strings.Split(url,matches),"/")
  }

  if (path == "") {
    err = errors.New("Bad Path ")
    return
  }

  switch (len(resq)) {
    case 2:
      quality,_ = strconv.Atoi(resq[1])
      res = strings.Split(resq[0],"x")
    case 1:
      res = strings.Split(resq[0],"x")
    default:
      log.Println("bad length ",len(resq))
  }

  if (res != nil) {
    width,_ = strconv.Atoi(res[0])
    height,_ = strconv.Atoi(res[1])
    if (width < 0 || height < 0) {
      err = errors.New("Invalid Aspect Ratio")
    }
  }
  return
}

/*
 This goroutine handles write to downstream. 
*/
func downstreamHandler(ds Downstream,ch chan DSData) {
  log.Println("Initializing downstream handler")
  for data := range ch {
    log.Println("received request for " + data.path)
    ds.Put(data)
  }
}

func Resizer(dwc DownstreamCfg, upc UpstreamCfg, scfg ServerCfg) (HandlerFunc) {

  var server Upstream
  var ds Downstream
  sizeRegex = regexp.MustCompile("/([0-9]+)x([0-9]+)/")
  chD := make(chan DSData)

  validateExtensions(scfg.Extensions)
  imagick.Initialize()

  url,err := url.Parse(upc.URI)
  if err != nil {
    log.Panic("Bad URL scheme")
  }

  switch url.Scheme {
    case "http":
      server = &HTTPUpstream{ upstreamURI: upc.URI}
      log.Println("Serving using " + upc.URI)
    case "file":
      server = &FileUpstream{ upstreamURI: url.Path }
      log.Println("Serving using " + url.Path)
    default:
      log.Panic("Unsupported url scheme " + url.Scheme)
  }

  if err = server.Init(upc); err != nil { 
      log.Panic("failed to initialize upstream")
  }

  if dwc.URI != "" {
    url,err = url.Parse(dwc.URI)

    if err != nil {
      log.Panic("Bad url scheme for downstream")
    }

    switch url.Scheme {
      case "s3":
        ds = &S3Downstream{ downstreamURI: dwc.URI }
        log.Println("Caching using " + url.Host)
      case "file":
        ds = &FileDownstream{ downstreamURI: url.Path}
        log.Println("Caching using " + url.Path)
      default:
        log.Panic("Unsupported downstream url scheme, disabling ", url.Scheme)
    }

    ds.Init()
    for i := 0; i < dwc.MaxThreads; i++ {
      go downstreamHandler(ds,chD)
    }
  }

  return func(w http.ResponseWriter, r* http.Request, next http.HandlerFunc) {

    var obuf []byte
    var wquality float32

    if (strings.HasPrefix(r.URL.Path,"/images/catalog/") == false) {
      log.Println("skipping ",r.URL.Path)
      next(w,r);
      return
    }

    err,filePath,width,height,quality := getFilePathResQuality(r.URL.Path)

    if err != nil || strings.LastIndex(filePath,".") == -1 {
      http.Error(w, "Forbidden", http.StatusForbidden)
      return
    }

    valid := scfg.ValidSizes

    if valid != "" && strings.Contains(valid,fmt.Sprintf("%dx%d",width,height)) != true {
      log.Printf("invalid size requested in %s, %dx%d\n",r.URL.Path,width,height);
      http.Error(w, "Invalid size specified.", http.StatusForbidden)
      return
    }

    // if .webp is at the end of url, webp has been requested
    ext := filePath[strings.LastIndex(filePath,"."):]

    if scfg.Extensions != "" && strings.Contains(scfg.Extensions,ext) != true {
      log.Printf("invalid extension %s requested in %s",ext,r.URL.Path)
      http.Error(w, "Unsupported Media", http.StatusUnsupportedMediaType)
      return
    }

    if ext == ".webp" {
      filePath = strings.TrimSuffix(filePath,".webp")
      wquality = float32(quality)
      quality = 100
    }
    
    start := time.Now()
    file,err := server.Get(w,r,filePath)

    if file != nil {
      defer file.Close() // in case of 404, file still needs to be closed.
    }

    log.Printf("completed fetch for %s in %v seconds\n",filePath,time.Since(start))

    if err != nil {
      log.Println("upstream error with ",r.URL.Path)
      http.Error(w, err.Error(), http.StatusNotFound)
      return
    }

    body, err := ioutil.ReadAll(file)

    if err != nil {
      log.Println("Failed to read image ", r.URL.Path)
      http.Error(w, err.Error(), http.StatusNotFound)
      return
    }

   if len(body) == 0 {
     log.Println("Bad Image size - zero content length ", r.URL.Path)
     http.Error(w, "Invalid Image", http.StatusBadRequest)
     return
   }

    mimeType := mime.TypeByExtension(ext)
    w.Header().Set("Content-Type", mimeType)
    w.Header().Set("Cache-Control","max-age=31556926")

    /*
    // for now, save original on downstream as well
    if (dwc.URI != "") {
      log.Println("issuing cache request for original ",filePath);
      chD <- DSData{data: &body, path: filePath, mimeType: mimeType}
    }
    */
 
    start = time.Now()
    if (width == 0 && height == 0) {
      obuf = body
    } else {
      obuf,err = Resize(uint(width), uint(height), uint(quality), body)
      body = nil // free reference
      if err != nil {
        log.Println("Failed to resize image ", r.URL.Path)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
      }
    }


    // if webp conversion was requested, convert to webp, after magicwand resize
    if ext == ".webp" {
      obuf,err = EncodeWebp(obuf,wquality)

      if err != nil {
        log.Println("failed to convert to webp ", filePath, err.Error())
        http.Error(w,err.Error(), http.StatusInternalServerError)
        return
      }
    }

    log.Println("completed resize in ",time.Since(start))

    w.Header().Set("Content-Length", strconv.FormatUint(uint64(len(obuf)), 10))
    w.WriteHeader(http.StatusOK)

    // cache the result, if we actually did a resize
    if (dwc.URI != "" && (width !=0 || height != 0) ) {
      log.Println("sending request to downstream for caching " + r.URL.Path)
      chD <- DSData{data: &obuf, path: r.URL.Path, mimeType: mimeType}
    }

    if r.Method != "HEAD" {
      w.Write(obuf)
    }

  }
}
