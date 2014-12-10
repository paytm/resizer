package resized

import (
  "code.google.com/p/gcfg"
  "log"
)

type UpstreamCfg struct {
  URI string
  Timeout string
}

type DownstreamCfg struct {
  URI string
  MaxThreads int
}

type ServerCfg struct {
   Port string
   ValidSizes string
   Extensions string
   Rate int
}

type Config struct {
  Upstream UpstreamCfg
  Downstream DownstreamCfg
  Server ServerCfg 
}

func ReadConfig(cfg *Config,path string) (ok bool) {
  err := gcfg.ReadFileInto(cfg,path + "/resizer.ini")
  if (err != nil) {
    log.Println("error reading file from ",path,err.Error())
    return false
  }
  return true
}
