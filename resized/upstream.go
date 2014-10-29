package resized

import (
  "net/http"
  "time"
  "io"
  "os"
  "errors"
  "log"
)

type Upstream interface {
  Init(UpstreamCfg) error
  Get(w http.ResponseWriter, r *http.Request, path string) (io.ReadCloser,error)
}

type FileUpstream struct {
  upstreamURI string
}

func (u *FileUpstream) Init(UpstreamCfg) error {
  return nil
}

func (u *FileUpstream) Get(w http.ResponseWriter, r *http.Request, path string) (file io.ReadCloser, err error) {
  file, err = os.Open(u.upstreamURI + path);
  return file,err
}

type HTTPUpstream struct {
  upstreamURI string
  client      *http.Client
}

func (u *HTTPUpstream) Init(upc UpstreamCfg) error {
  d,err := time.ParseDuration(upc.Timeout)
  if err == nil {
    u.client = &http.Client{ Timeout: d }
    log.Println("created client with timeout ",d);
  }
  return err
}

func (u *HTTPUpstream) Get(w http.ResponseWriter, r *http.Request, path string) (file io.ReadCloser, err error) {
    log.Println("fetching " + u.upstreamURI + path)
    resp,err := u.client.Get(u.upstreamURI + path)

    if (err == nil) {
      file = resp.Body
      if (resp.StatusCode != 200)  {
        err = errors.New("Not Found")
      }
    }

    return file,err
}
