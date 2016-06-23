package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/paytm/logging"
	config "github.com/qzaidi/consulcfg"
	grace "gopkg.in/paytm/grace.v1"
	"paytm.unfuddle.com/resizer/resized"
	//"github.com/paytm/resizer/middleware"
)

func main() {

	var cfg resized.Config

	cfgpath := flag.String("c", "./resizer.ini", "config file path")

	v := flag.Bool("version", false, "prints resizer version")
	useConsul := flag.Bool("consul", false, "use consul for config")

	flag.Parse()

	if *v {
		fmt.Println(ResizerVersion())
		os.Exit(0)
	}

	var ok bool
	if *useConsul {
		ok = config.ReadConfig("resizer", &cfg)
	} else {
		log.Println("using config from ", *cfgpath)
		ok = resized.ReadConfig(&cfg, *cfgpath) || resized.ReadConfig(&cfg, "/etc/resizer.ini")
	}

	if !ok {
		log.Println("failed to read resizer config")
		os.Exit(1)
	}

	logging.LogInit()

	/* QZ: disable rate limits for now, first we move to graceful
	if cfg.Server.Rate != 0 {
	  fmt.Printf("Rate limiting at %d req/sec\n",cfg.Server.Rate)
	  n.Use(middleware.Ratelimit(cfg.Server.Rate))
	}
	*/

	http.Handle("/", resized.Resizer(cfg.Downstream, cfg.Upstream, cfg.Server))
	log.Fatal(grace.Serve(":"+cfg.Server.Port, nil))

}
