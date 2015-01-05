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

func ReadConfig(cfg *Config,path string) (ok bool) {
  err := gcfg.ReadFileInto(cfg,path)
  if (err != nil) {
    return false
  }
  return true
}
