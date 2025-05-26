package http

import "net/http"

var prefix = "/api"
var Router = map[string]func(http.ResponseWriter, *http.Request){
	"/":      helloHandler,
	"/param": paramHandler,
}
