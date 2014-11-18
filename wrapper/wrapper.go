package wrapper

import (
    "net/http"
    "../ratelimit"
    "../resized"
    )

// Ratelimit and resized have to be pointed to github repo

func RateLimit(h resized.HandlerFunc, rate int) (resized.HandlerFunc) {
    rl, _ := ratelimit.NewRateLimiter(rate)

    return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc){
        if rl.Limit(){
            http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout )
        } else {
            next(w, r)
        }
    }
}
