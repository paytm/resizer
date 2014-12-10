package middleware

import (
    "net/http"
    "github.com/paytm/resizer/ratelimit"
    )

// Ratelimit has to be pointed to github repo

type RatelimitStruct struct {
    limiter *ratelimit.Ratelimiter
}

// A struct that has a ServeHTTP method
func Ratelimit(rate  int) *RatelimitStruct{
    rl, _  := ratelimit.NewRateLimiter(rate)
    return &RatelimitStruct{rl}
}

func (r *RatelimitStruct) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc){
    if r.limiter.Limit(){
        http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout )
    } else {
        next(w, req)
    }
}
