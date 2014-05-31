package resizer

import (
  "net/http"
  "io"
  "os"
  "errors"
  "log"
)

type Upstream interface {
  ServeOriginal(w http.ResponseWriter, r *http.Request, path string)
  Get(w http.ResponseWriter, r *http.Request, path string) (io.ReadCloser,error)
}

type FileUpstream struct {
  upstreamURI string
}

func (u *FileUpstream) ServeOriginal(w http.ResponseWriter, r*http.Request, path string) {
  http.ServeFile(w,r,u.upstreamURI + path)
}

func (u *FileUpstream) Get(w http.ResponseWriter, r *http.Request, path string) (file io.ReadCloser, err error) {
  file, err = os.Open(u.upstreamURI + path);
  return file,err
}

type HTTPUpstream struct {
  upstreamURI string
}

func (u *HTTPUpstream) ServeOriginal(w http.ResponseWriter, r*http.Request, path string) {
  log.Println("serving ", u.upstreamURI + path)
  http.Redirect(w,r,u.upstreamURI + path,302)
}

func (u *HTTPUpstream) Get(w http.ResponseWriter, r *http.Request, path string) (file io.ReadCloser, err error) {
    log.Println("fetching " + u.upstreamURI + path)
    resp,err := http.Get(u.upstreamURI + path)

    if (err == nil) {
      file = resp.Body
      if (resp.StatusCode != 200)  {
        err = errors.New("Not Found")
      }
    }

    return file,err
}
