package wrapper

import (
    "net/http"
    "../ratelimit"
    "../resized"
    "log"
    "os"
    )

// Ratelimit and resized have to be pointed to github repo

func RateLimit(h resized.HandlerFunc) (resized.HandlerFunc) {
    var cfg resized.Config
    ok := resized.ReadConfig(&cfg, "..") || resized.ReadConfig(&cfg,"/etc")
    if (!ok) {
        log.Println("failed to read resizer.ini from CWD or /etc")
        os.Exit(1)
    }

    rl, _ := ratelimit.NewRateLimiter(cfg.Server.Rate)

    return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc){
        if rl.Limit(){
            http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout )
        } else {
            next(w, r)
        }
    }
}