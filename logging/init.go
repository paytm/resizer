// This package handles log rotation
package logging

import (
  "os"
  "syscall"
  "flag"
  "log"
  "os/signal"
)

var stdoutLog string
var stderrLog string

func Init() {
  flag.StringVar(&stdoutLog,"l","","log file for stdout")
  flag.StringVar(&stderrLog,"e","","log file for stderr")

  c := make(chan os.Signal, 1)
  signal.Notify(c, syscall.SIGHUP) // listen for sighup
  go sigHandler(c)
}

func sigHandler(c chan os.Signal) {
  // Block until a signal is received.
  for s := range c {
    log.Println("Reloading on :", s)
    LogInit()
  }
}

func LogInit() {
  log.Println("Log Init: using ",stdoutLog,stderrLog)
  reopen(1,stdoutLog)
  reopen(2,stderrLog)
}

func reopen(fd int,filename string) {
  if filename == "" {
    return
  }

  logFile,err := os.OpenFile(filename, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0644)

  if (err != nil) {
    log.Println("Error in opening ",filename,err)
    return
  }

  syscall.Dup2(int(logFile.Fd()), fd)
}
