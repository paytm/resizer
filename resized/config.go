package resized

import (
  "code.google.com/p/gcfg"
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

func ReadConfig(cfg *Config,path string) bool {
  err := gcfg.ReadFileInto(cfg,path + "/resizer.ini")
  if err == nil {
    return true
  }
  return false
}
