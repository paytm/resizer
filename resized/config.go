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

type Config struct {
  Upstream UpstreamCfg
  Downstream DownstreamCfg
  Server struct {
   Port string
   ValidSizes string
  }
}

func ReadConfig(cfg *Config,path string) bool {
  err := gcfg.ReadFileInto(cfg,path + "/resizer.ini")
  if err == nil {
    return true
  }
  return false
}
