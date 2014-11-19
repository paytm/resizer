package middleware

import (
    "net/http"
    "../ratelimit"
    )

// Ratelimit has to be pointed to github repo

type RatelimitStruct struct {
    rate int
}

// A struct that has a ServeHTTP method
func Ratelimit(rate  int) *RatelimitStruct{
    return &RatelimitStruct{rate}
}

func (r *RatelimitStruct) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc){
    rl, _ := ratelimit.NewRateLimiter(r.rate)

    if rl.Limit(){
        http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout )
    } else {
        next(w, req)
    }
}
