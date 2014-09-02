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
  "github.com/chai2010/webp"
  "log"
  "strings"
  "strconv"
  "mime"
  "errors"
  "io/ioutil"
  "regexp"
  "bytes"
  "time"
  "fmt"
  "image"
  _ "image/jpeg"
  _ "image/png"
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

func Resizer(dws string, numDSThreads int, ups string,valid string) (HandlerFunc) {

  var server Upstream
  var ds Downstream
  sizeRegex = regexp.MustCompile("/([0-9]+)x([0-9]+)/")
  chD := make(chan DSData)

  imagick.Initialize()

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

  if dws != "" {
    url,err = url.Parse(dws)

    if err != nil {
      log.Panic("Bad url scheme for downstream")
    }

    switch url.Scheme {
      case "s3":
        ds = &S3Downstream{ downstreamURI: dws }
        log.Println("Caching using " + url.Host)
      case "file":
        ds = &FileDownstream{ downstreamURI: url.Path}
        log.Println("Caching using " + url.Path)
      default:
        log.Panic("Unsupported downstream url scheme, disabling ", url.Scheme)
    }

    ds.Init()
    for i := 0; i < numDSThreads; i++ {
      go downstreamHandler(ds,chD)
    }
  }

  return func(w http.ResponseWriter, r* http.Request, next http.HandlerFunc) {

    var obuf []byte

    if (strings.HasPrefix(r.URL.Path,"/images/catalog/") == false) {
      log.Println("skipping ",r.URL.Path)
      next(w,r);
      return
    }

    err,filePath,width,height,quality := getFilePathResQuality(r.URL.Path)

    if err != nil {
      http.Error(w, err.Error(), http.StatusForbidden)
      return
    }

    if valid != "" && strings.Contains(valid,fmt.Sprintf("%dx%d",width,height)) != true {
      log.Printf("invalid size requested in %s, %dx%d\n",r.URL.Path,width,height);
      http.Error(w, "Invalid size specified.", http.StatusForbidden)
      return
    }

    // if .webp is at the end of url, webp has been requested
    ext := filePath[strings.LastIndex(filePath,"."):]

    if ext == ".webp" {
      filePath = strings.TrimSuffix(filePath,".webp")
      wquality := quality
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
      http.Error(w, "File not found", http.StatusNotFound)
      return
    }

    body, err := ioutil.ReadAll(file)

    if err != nil {
      log.Println("Failed to read image ", r.URL.Path)
      http.Error(w, err.Error(), http.StatusNotFound)
      return
    }


    mimeType := mime.TypeByExtension(filePath[strings.LastIndex(filePath,"."):])
    w.Header().Set("Content-Type", mimeType)

    // for now, save original on downstream as well
    if (dws != "") {
      log.Println("issuing cache request for original ",filePath);
      chD <- DSData{data: &body, path: filePath, mimeType: mimeType}
    }
 
    start = time.Now()
    if (width == 0 && height == 0) {
      obuf = body
    } else {
      obuf,err = Resize(uint(width), uint(height), uint(quality), body)
      if err != nil {
        log.Println("Failed to resize image ", r.URL.Path)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
      }
    }


    // if webp conversion was requested, convert to webp, after magicwand resize
    if ext == ".webp" {
      var data bytes.Buffer
      src := bytes.NewBuffer(obuf)

      img,_,err := image.Decode(src)

      if err != nil {
        log.Println("failed to decode magickwand output for ", filePath)
        http.Error(w,err.Error(), http.StatusNotFound)
        return
      }

      if webp.Encode(&data, img, &webp.Options{ false, unit(wquality)}); err != nil {
        log.Println("failed to convert image to webp for ", filePath)
        http.Error(w,err.Error(), http.StatusNotFound)
        return
      }

      obuf = data.Bytes()
    }

    log.Println("completed resize in ",time.Since(start))

    w.Header().Set("Content-Length", strconv.FormatUint(uint64(len(obuf)), 10))
    w.WriteHeader(http.StatusOK)

    // cache the result, if we actually did a resize
    if (dws != "" && (width !=0 || height != 0) ) {
      log.Println("sending request to downstream for caching " + r.URL.Path)
      chD <- DSData{data: &obuf, path: r.URL.Path, mimeType: mimeType}
    }

    if r.Method != "HEAD" {
      w.Write(obuf)
    }

  }
}
