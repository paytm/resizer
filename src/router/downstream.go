package router

import (
  "log"
  "os"
  "path"
  "net/url"
  "github.com/crowdmob/goamz/aws"
  "github.com/crowdmob/goamz/s3"
)

type DSData struct {
  data *[]byte
  path string
  mimeType string
}

type Downstream interface {
  Init() error
  Put(DSData) error
}

type FileDownstream struct {
  downstreamURI string
}

func (d *FileDownstream) Init() error {
  // TODO: check to see if downstreamURI is valid
  return nil
}

func (d *FileDownstream) Put (data DSData) error {
  cachePath := d.downstreamURI + data.path
  err := os.MkdirAll(path.Dir(cachePath),os.ModeDir | 0777)
  if err == nil {
    out, _ := os.Create(cachePath)
    out.Write(*data.data)
    out.Close()
    log.Println("cached into " + cachePath);
  } else {
    log.Println("cache fail ",err.Error())
  }
  return err
}


type S3Downstream struct {
  downstreamURI string
  bucket *s3.Bucket
}

func (d *S3Downstream) Init() error {
  u,err := url.Parse(d.downstreamURI)
  if err != nil || u.Scheme != "s3" {
    log.Panic("Bad URL scheme ",d.downstreamURI)
  }

  username := u.User.Username()
  password,_ := u.User.Password()

  auth := aws.Auth { AccessKey: username, SecretKey: password }

  log.Println("Init s3 connection using key ", auth.AccessKey,u.Host)
  connection := s3.New(auth,aws.APSoutheast)
  d.bucket = connection.Bucket(u.Host)
  return nil
}

func (d *S3Downstream) Put (data DSData) error {
  // first do a get
  ok, _ := d.bucket.Exists(data.path)

  if ok == true {
    log.Println("file already exists at s3 ",data.path)
    return nil
  }

  err := d.bucket.Put(data.path,*data.data,data.mimeType,s3.PublicRead,s3.Options{})
  if err == nil {
    log.Println("saved on s3 ", data.path)
  }
  return err
}
